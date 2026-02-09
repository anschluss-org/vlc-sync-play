package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
)

type Syncer struct {
	syncingMu                    sync.Mutex
	players                      *players
	settings                     Settings
	followersSkipUpdatesDuration time.Duration
	state                        State
	isStarted                    atomic.Bool
	instanceLauncher             instance.Launcher
	logger                       logging.Logger
	filePaths                    []string // File paths for each instance
}

func NewSyncer(
	settings Settings,
	instanceLauncher instance.Launcher,
	logger logging.Logger,
	filePaths []string,
) *Syncer {
	return &Syncer{
		players:  newPlayers(),
		settings: settings,
		followersSkipUpdatesDuration: timings.GetFollowerUpdatesIgnoreDuration(
			settings.GetPollingInterval().GetValue(),
		),
		state:            NewState(),
		instanceLauncher: instanceLauncher,
		logger:           logger,
		filePaths:        filePaths,
	}
}

func (s *Syncer) Start(ctx context.Context, initFileURI string) error {
	// Launch all instances at startup (not just 1)
	instancesNumber := s.settings.GetInstancesNumber().GetValue()
	s.logger.Info("=== Starting with %d instances, initFileURI: %s ===", instancesNumber, initFileURI)
	s.logger.Info("=== filePaths configured: %v ===", s.filePaths)
	if err := s.launchInstances(ctx, initFileURI, instancesNumber); err != nil {
		return err
	}
	s.logger.Info("=== All instances launched ===")

	// Wait a moment, then check all instance states
	go func() {
		time.Sleep(2 * time.Second)
		s.logger.Info("=== Checking initial instance states ===")
		s.players.Iterate(func(pl *player) bool {
			status, err := pl.client.client.GetStatusEx(ctx, repetition.Single())
			if err != nil {
				s.logger.Err("P[%d]: Failed to get status: %s", pl.GetID(), err.Error())
			} else {
				s.logger.Info("P[%d]: FileURI=%s, State=%s, Position=%.2f, Length=%.2f",
					pl.GetID(), status.FileURI, status.State, status.Position, status.LengthSec)
			}
			return true
		})
		s.logger.Info("=== Initial state check complete ===")
	}()

	ctx, cancel := context.WithCancel(ctx)

	defer s.settings.GetInstancesNumber().Subscribe(func(value int) {
		s.launchMissingInstances(ctx, value)
	}).Unsubscribe()

	defer s.settings.GetPollingInterval().Subscribe(func(value time.Duration) {
		s.syncingMu.Lock()
		defer s.syncingMu.Unlock()
		s.followersSkipUpdatesDuration = timings.GetFollowerUpdatesIgnoreDuration(value)
	})

	return s.players.WaitAndPoll(
		ctx,
		func(update playerUpdate) {
			if err := s.onUpdate(ctx, &update); err != nil {
				cancel()
			}
		},
		func(event playerEvent) {
			if err := s.onEvent(ctx, event); err != nil {
				cancel()
			}
		},
		s.onFinished,
	)
}

func (s *Syncer) onEvent(ctx context.Context, event playerEvent) error {
	switch event.event {
	case instance.StderrEventMouse1Click:
		// Mouse click detected - sync all players to clicked player's state
		go func() {
			// Small delay to let the click action complete
			time.Sleep(150 * time.Millisecond)

			s.logger.Info("Mouse click detected in P[%d], syncing all players", event.player.GetID())

			s.syncingMu.Lock()
			defer s.syncingMu.Unlock()

			// Get current status of the clicked player
			status, err := event.player.client.client.GetStatusEx(ctx, repetition.Single())
			if err != nil {
				s.logger.Err("Failed to get status after click: %s", err.Error())
				return
			}

			// Create update to reflect the state change
			update, err := event.player.client.state.GetUpdate(&status)
			if err != nil {
				return
			}

			// Mark this as the source and trigger sync
			s.state.lastSyncedFromID = event.player.GetID()
			plUpdate := &playerUpdate{
				player: event.player,
				update: update,
			}

			// Sync the state change (this will sync play/pause state)
			if update.ChangedProps.HasState() || update.ChangedProps.HasPosition() {
				s.syncPlayers(ctx, plUpdate)
			}
		}()
	}
	return nil
}

func (s *Syncer) onFinished(pl *player) {
	s.logger.Info("P[%d]: finished", pl.GetID())
	s.syncingMu.Lock()
	defer s.syncingMu.Unlock()
	if s.state.lastSyncedFromID == pl.GetID() {
		s.state.lastSyncedFromID = instance.IDNone
	}
}

func (s *Syncer) sendAllPlayersCommands(
	ctx context.Context,
	commands extended.CmdGroup,
) {
	noSeekCommands := commands
	noSeekCommands.Seek.Reset()

	waitGr := sync.WaitGroup{}
	s.players.Iterate(func(pl *player) bool {
		waitGr.Add(1)
		go func() {
			defer waitGr.Done()
			_, _ = pl.SendCmdGroup(ctx, commands, repetition.WithInterval(timings.CommandsRepeatInterval))
		}()
		return true
	})
	waitGr.Wait()

	s.syncPlayersPosition(ctx, commands.Seek.Value, nil)
}

func (s *Syncer) onUpdate(ctx context.Context, plUpdate *playerUpdate) error {
	s.syncingMu.Lock()
	defer s.syncingMu.Unlock()

	s.logger.Info(">>> P[%d] UPDATE: State=%s, Position=%.2f, IsNatural=%v, Changed=%s",
		plUpdate.player.GetID(), plUpdate.update.Status.State, plUpdate.update.Status.Position,
		plUpdate.update.IsNatural, plUpdate.update.ChangedProps.String())

	if canAccept, getReason := s.checkCanAcceptUpdate(plUpdate); !canAccept {
		s.logger.Info(getReason())
		return nil
	}

	s.state.fileURI.SetValue(plUpdate.update.Status.FileURI)
	s.state.lastSyncedFromID = plUpdate.player.GetID()

	if plUpdate.update.ChangedProps.HasFileURI() &&
		plUpdate.update.Status.State != basic.PlaybackStateStopped {
		s.logger.Info(">>> File opened event detected")
		s.onFileOpened(ctx, plUpdate.player.GetID())
		return nil
	}

	if !plUpdate.update.IsNatural {
		s.logger.Info(">>> Non-natural update, triggering sync")
		s.syncPlayers(ctx, plUpdate)
	} else if plUpdate.update.ChangedProps.HasState() {
		// Natural state change (user clicked play/pause)
		// Sync to other instances if this is the source player
		if plUpdate.player.GetID() == s.state.lastSyncedFromID || s.state.lastSyncedFromID == instance.IDNone {
			s.logger.Info(">>> Natural state change from source player P[%d]: %s -> triggering sync", plUpdate.player.GetID(), plUpdate.update.Status.State)
			s.syncPlayers(ctx, plUpdate)
		} else {
			s.logger.Info(">>> Natural state change from non-source player P[%d] (source is P[%d]) - NOT syncing", plUpdate.player.GetID(), s.state.lastSyncedFromID)
		}
	} else {
		s.logger.Info(">>> Update doesn't trigger sync (no state change or not natural)")
	}
	return nil
}

func (s *Syncer) checkCanAcceptUpdate(plUpdate *playerUpdate) (canAccept bool, getReason func() string) {
	plID := plUpdate.player.GetID()
	if plID == s.state.lastSyncedFromID || s.state.lastSyncedFromID == instance.IDNone {
		return true, nil
	}

	if plUpdate.update.Status.Moment.Center().Before(s.state.acceptFollowerUpdatesAfter) {
		return false, func() string {
			return fmt.Sprintf("Skipping [%d] update from %v old sync iteration: pos %v",
				plUpdate.player.GetID(),
				s.state.acceptFollowerUpdatesAfter.Sub(plUpdate.update.Status.Moment.Max),
				plUpdate.update.Status.Position,
			)
		}
	}
	return true, nil
}

func (s *Syncer) launchMissingInstances(ctx context.Context, targetInstancesNumber int) {
	missing := targetInstancesNumber - s.players.Len()
	if missing <= 0 {
		return
	}

	if fileURI := s.state.fileURI.GetValue(); fileURI != "" {
		_ = s.launchInstances(ctx, fileURI, missing)
	} else {
		defer s.state.fileURI.Subscribe(func(fileURI string) {
			_ = s.launchInstances(ctx, fileURI, missing)
		}).Unsubscribe()
	}
	return
}

func (s *Syncer) launchInstances(
	ctx context.Context,
	fileURI string,
	missingInstancesNumber int,
) error {
	currentPlayerCount := s.players.Len()

	// Launch instances SEQUENTIALLY to ensure correct file order
	for i := 0; i < missingInstancesNumber; i++ {
		// Calculate which instance number this will be (0-based for array access)
		instanceIndex := currentPlayerCount + i

		// Get the appropriate file for this instance
		var instanceFileURI string
		if len(s.filePaths) > instanceIndex {
			instanceFileURI = s.filePaths[instanceIndex]
			s.logger.Info("Instance %d will use file: %s", instanceIndex+1, instanceFileURI)
		} else if len(s.filePaths) > 0 {
			instanceFileURI = s.filePaths[0]
			s.logger.Info("Instance %d will use fallback file: %s", instanceIndex+1, instanceFileURI)
		} else {
			instanceFileURI = fileURI
		}

		options := instance.LaunchOptions{
			// First instance will be launched with video
			NoVideo: instanceIndex > 0 && s.settings.GetNoVideo().GetValue(),
			FileURI: typeutil.Optional[string]{
				HasValue: instanceFileURI != "",
				Value:    instanceFileURI,
			},
		}

		s.logger.Info("Launching new instance with file: %s", instanceFileURI)

		newInstance, err := s.instanceLauncher.Launch(ctx, options)
		if err != nil {
			return fmt.Errorf("failed to create new instance: %w", err)
		}

		newPl := newPlayer(
			newInstance,
			getPlayerSettings(s.settings),
			s.logger,
		)
		s.players.Add(newPl)

		s.logger.Info("Instance launched with ID P[%d], file: %s", newPl.GetID(), instanceFileURI)
	}
	return nil
}

func (s *Syncer) onFileOpened(ctx context.Context, srcPlayerID uint) {
	// The source player may auto-seek and will send the next update with other properties
	s.state.acceptFollowerUpdatesAfter = time.Now().Add(max(
		s.followersSkipUpdatesDuration,
		timings.WaitForAutoSeekAfterFileOpenedDuration,
	))

	// Launch any missing instances
	go func() {
		s.launchMissingInstances(ctx, s.settings.GetInstancesNumber().GetValue())

		// If launched without files, try to open the paired file in other instances
		if len(s.filePaths) == 0 {
			s.logger.Info("No initial files configured - attempting to find paired file")
			s.openPairedFileInOtherInstances(ctx, srcPlayerID)
		}

		// After launching, wait for new instances to be ready, then sync them
		time.Sleep(300 * time.Millisecond)
		s.syncNewInstancesWithSource(ctx, srcPlayerID)
	}()
}

// openPairedFileInOtherInstances finds the paired file and opens it in other instances
func (s *Syncer) openPairedFileInOtherInstances(ctx context.Context, srcPlayerID uint) {
	// Find the source player
	var srcPlayer *player
	s.players.Iterate(func(pl *player) bool {
		if pl.GetID() == srcPlayerID {
			srcPlayer = pl
			return false
		}
		return true
	})

	if srcPlayer == nil {
		return
	}

	// Get source player's file
	srcStatus, err := srcPlayer.client.client.GetStatusEx(ctx, repetition.Single())
	if err != nil || srcStatus.FileURI == "" {
		s.logger.Err("Cannot find paired file - source has no file loaded")
		return
	}

	// Parse the file URI (remove file:// prefix if present)
	srcFile := strings.TrimPrefix(srcStatus.FileURI, "file://")

	s.logger.Info("Source file: %s", srcFile)

	// Find the paired file
	pairedFile := s.findPairedFile(srcFile)
	if pairedFile == "" {
		s.logger.Info("No paired file found for: %s", srcFile)
		return
	}

	s.logger.Info("Found paired file: %s", pairedFile)

	// Open the paired file in all other instances
	wg := sync.WaitGroup{}
	s.players.Iterate(func(pl *player) bool {
		if pl.GetID() != srcPlayerID {
			wg.Add(1)
			go func(player *player) {
				defer wg.Done()

				s.logger.Info("P[%d]: Opening paired file: %s", player.GetID(), pairedFile)

				// Open file
				_, err := player.client.client.SendCmdGroup(
					ctx,
					extended.CmdGroup{
						OpenFile: typeutil.NewOptional(pairedFile),
					},
					repetition.WithInterval(timings.CommandsRepeatInterval),
				)

				if err != nil {
					s.logger.Err("P[%d]: Failed to open paired file: %s", player.GetID(), err.Error())
					return
				}

				// Wait for playback to actually start before pausing
				s.logger.Info("P[%d]: Waiting for playback to start...", player.GetID())
				maxAttempts := 50 // 5 seconds max
				for attempt := 0; attempt < maxAttempts; attempt++ {
					status, err := player.client.client.GetStatusEx(ctx, repetition.Single())
					if err == nil && status.State == basic.PlaybackStatePlaying && status.LengthSec > 0 {
						s.logger.Info("P[%d]: Playback started, now pausing", player.GetID())
						break
					}
					time.Sleep(100 * time.Millisecond)
				}

				// Pause this instance
				_, _ = player.client.client.SendCmdGroup(
					ctx,
					extended.CmdGroup{
						State: typeutil.NewOptional(basic.PlaybackStatePaused),
					},
					repetition.WithInterval(100 * time.Millisecond),
				)

				s.logger.Info("P[%d]: Paired file opened and paused", player.GetID())
			}(pl)
		}
		return true
	})
	wg.Wait()

	// Also pause the source player to keep both instances paused
	s.logger.Info("P[%d]: Pausing source player", srcPlayerID)
	if srcPlayer != nil {
		_, _ = srcPlayer.client.client.SendCmdGroup(
			ctx,
			extended.CmdGroup{
				State: typeutil.NewOptional(basic.PlaybackStatePaused),
			},
			repetition.WithInterval(100 * time.Millisecond),
		)
		s.logger.Info("P[%d]: Source player paused", srcPlayerID)
	}
}

// findPairedFile attempts to find the paired file based on naming convention
func (s *Syncer) findPairedFile(fileURI string) string {
	// Parse the file path
	dir := filepath.Dir(fileURI)
	filename := filepath.Base(fileURI)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	var pairedFile string

	// Check if it's a conductor file
	if strings.HasSuffix(nameWithoutExt, "-conductor") {
		basename := strings.TrimSuffix(nameWithoutExt, "-conductor")
		pairedFile = filepath.Join(dir, basename+"-audience"+ext)
	} else if strings.HasSuffix(nameWithoutExt, "-audience") {
		basename := strings.TrimSuffix(nameWithoutExt, "-audience")
		pairedFile = filepath.Join(dir, basename+"-conductor"+ext)
	} else {
		// No convention match - try to find conductor or audience variant
		conductorFile := filepath.Join(dir, nameWithoutExt+"-conductor"+ext)
		audienceFile := filepath.Join(dir, nameWithoutExt+"-audience"+ext)

		// Check which one exists
		if _, err := os.Stat(conductorFile); err == nil {
			pairedFile = conductorFile
		} else if _, err := os.Stat(audienceFile); err == nil {
			pairedFile = audienceFile
		}
	}

	// Verify the paired file exists
	if pairedFile != "" {
		if _, err := os.Stat(pairedFile); err != nil {
			s.logger.Info("Paired file does not exist: %s", pairedFile)
			return ""
		}
	}

	return pairedFile
}

// syncNewInstancesWithSource syncs newly launched instances with the source player's current state
func (s *Syncer) syncNewInstancesWithSource(ctx context.Context, srcPlayerID uint) {
	// Find the source player
	var srcPlayer *player
	s.players.Iterate(func(pl *player) bool {
		if pl.GetID() == srcPlayerID {
			srcPlayer = pl
			return false
		}
		return true
	})

	if srcPlayer == nil {
		return
	}

	// Get source player's current status
	srcStatus, err := srcPlayer.client.client.GetStatusEx(ctx, repetition.Single())
	if err != nil {
		s.logger.Err("Failed to get source status for sync: %s", err.Error())
		return
	}

	// Only sync if source is actually playing
	if srcStatus.State != basic.PlaybackStatePlaying {
		s.logger.Info("Source player P[%d] not playing, skipping sync", srcPlayerID)
		return
	}

	s.logger.Info("Syncing new instances with playing source player P[%d]", srcPlayerID)

	// Build commands to match source state
	commands := extended.CmdGroup{
		State: typeutil.NewOptional(basic.PlaybackStatePlaying),
	}

	// Also sync position if available
	if srcStatus.Position > 0 {
		expectedPos := srcStatus.Position
		commands.Seek.Set(func(atMoment time.Time) float64 {
			return expectedPos
		})
	}

	// Send commands to all other players that have actually loaded a file
	wg := sync.WaitGroup{}
	s.players.Iterate(func(pl *player) bool {
		if pl.GetID() != srcPlayerID {
			wg.Add(1)
			go func(player *player) {
				defer wg.Done()

				// Wait for file to load with timeout
				timeout := time.After(10 * time.Second)
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()

				fileLoaded := false
				for !fileLoaded {
					select {
					case <-ctx.Done():
						s.logger.Info("P[%d]: Context cancelled while waiting for file", player.GetID())
						return
					case <-timeout:
						s.logger.Err("P[%d]: Timeout waiting for file to load", player.GetID())
						return
					case <-ticker.C:
						status, err := player.client.client.GetStatusEx(ctx, repetition.Single())
						if err != nil {
							s.logger.Err("P[%d]: Failed to get status: %s", player.GetID(), err.Error())
							continue
						}

						// Check if file is loaded (FileURI not empty and length > 0)
						if status.FileURI != "" && status.LengthSec > 0 {
							fileLoaded = true
							s.logger.Info("P[%d]: File loaded, ready to sync", player.GetID())
						}
					}
				}

				s.logger.Info("P[%d]: Sending initial play command", player.GetID())
				_, _ = player.SendCmdGroup(ctx, commands, repetition.WithInterval(timings.CommandsRepeatInterval))
			}(pl)
		}
		return true
	})
	wg.Wait()
}

// getFileURIForInstance returns the file path for a specific instance ID
// Instance IDs start at 1, so we subtract 1 to get the array index
func (s *Syncer) getFileURIForInstance(instanceID uint) string {
	if len(s.filePaths) == 0 {
		return s.state.fileURI.GetValue()
	}

	// Convert instance ID to 0-based index
	index := int(instanceID) - 1

	// If we have a specific file for this instance, use it
	if index >= 0 && index < len(s.filePaths) {
		return s.filePaths[index]
	}

	// Fall back to the first file if we don't have enough files
	return s.filePaths[0]
}

func (s *Syncer) syncPlayers(
	ctx context.Context,
	srcUpdate *playerUpdate,
) {
	s.logger.Info("-- Syncing caused by %d update: %s", srcUpdate.player.GetID(), srcUpdate.update.String())

	// Don't sync FileURI - each instance should keep its own file
	// Files are set during launch or explicit file open events only
	syncProps := srcUpdate.update.ChangedProps
	syncProps.SetFileURI(false)

	commands := srcUpdate.player.client.state.GetSyncCommands(syncProps)
	s.syncOtherPlayersNoSeek(ctx, srcUpdate, commands)
	if commands.Seek.HasValue && s.players.Len() > 1 {
		var skipPlayer *player
		if !s.settings.GetReSeekSrc().GetValue() {
			skipPlayer = srcUpdate.player
		}
		s.syncPlayersPosition(ctx, commands.Seek.Value, skipPlayer)
	}
	s.state.acceptFollowerUpdatesAfter = time.Now().Add(s.followersSkipUpdatesDuration)
}

func (s *Syncer) syncOtherPlayersNoSeek(
	ctx context.Context,
	srcUpdate *playerUpdate,
	commands extended.CmdGroup,
) {
	commands.Seek.Reset()
	srcUpdate.update.ChangedProps.SetPosition(false)
	if !commands.HasAny() {
		return
	}

	waitGr := sync.WaitGroup{}

	s.players.Iterate(func(pl *player) bool {
		dstCommands := commands
		if srcUpdate.player == pl {
			return true
		}

		// Check if additional props sync required
		dstUpdate, err := pl.client.state.GetUpdate(&srcUpdate.update.Status)
		dstUpdate.ChangedProps.SetPosition(false)
		dstUpdate.ChangedProps.SetFileURI(false) // Don't sync FileURI - each instance has its own file

		if err == nil && !srcUpdate.update.ChangedProps.Includes(dstUpdate.ChangedProps) {
			s.logger.Info("P[%d]: additional sync [%s] -> [%s]", pl.GetID(), srcUpdate.update, dstUpdate)
			additionalProps := srcUpdate.update.ChangedProps.Union(dstUpdate.ChangedProps)
			additionalProps.SetFileURI(false) // Don't sync FileURI
			dstCommands = srcUpdate.player.client.state.GetSyncCommands(additionalProps)
		}

		waitGr.Add(1)
		go func() {
			defer waitGr.Done()
			_, _ = pl.SendCmdGroup(ctx, dstCommands, repetition.WithInterval(timings.CommandsRepeatInterval))
		}()
		return true
	})
	waitGr.Wait()
}

func (s *Syncer) syncPlayersPosition(
	ctx context.Context,
	positionGetter extended.ExpectedPositionGetter,
	skipPlayer *player,
) {
	commands := extended.CmdGroup{
		Seek: typeutil.NewOptional(positionGetter),
	}

	// Single attempt, no retry loop to avoid choppy audio
	wg := sync.WaitGroup{}

	s.players.Iterate(func(pl *player) bool {
		if pl == skipPlayer {
			return true
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := pl.SendCmdGroup(ctx, commands, repetition.Single()); err != nil {
				// Log but don't retry - reduces audio choppiness
				s.logger.Info("Position sync attempt failed (will retry on next update): %s", err.Error())
			}
		}()
		return true
	})
	wg.Wait()
}
