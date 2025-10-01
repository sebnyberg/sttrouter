package clipboard

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

// CopyToClipboard copies data from the provided io.Reader to the system clipboard
// Uses wl-copy for Wayland or xclip for X11
func CopyToClipboard(ctx context.Context, log *slog.Logger, reader io.Reader) error {
	// Try wl-copy first (Wayland)
	if _, err := exec.LookPath("wl-copy"); err == nil {
		log.DebugContext(ctx, "Copying to clipboard using wl-copy")
		cmd := exec.CommandContext(ctx, "wl-copy")
		cmd.Stdin = reader

		if err := cmd.Run(); err != nil {
			log.ErrorContext(ctx, "Failed to copy to clipboard with wl-copy", "error", err)
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}

		log.DebugContext(ctx, "Successfully copied to clipboard")
		return nil
	}

	// Try xclip (X11)
	if _, err := exec.LookPath("xclip"); err == nil {
		log.DebugContext(ctx, "Copying to clipboard using xclip")
		cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard")
		cmd.Stdin = reader

		if err := cmd.Run(); err != nil {
			log.ErrorContext(ctx, "Failed to copy to clipboard with xclip", "error", err)
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}

		log.DebugContext(ctx, "Successfully copied to clipboard")
		return nil
	}

	return fmt.Errorf("no clipboard tool found: install wl-copy (wayland) or xclip (X11)")
}
