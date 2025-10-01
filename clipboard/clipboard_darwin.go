package clipboard

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

// CopyToClipboard copies data from the provided io.Reader to the system clipboard using pbcopy
func CopyToClipboard(ctx context.Context, log *slog.Logger, reader io.Reader) error {
	log.DebugContext(ctx, "Copying to clipboard using pbcopy")

	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = reader

	// Run the command
	if err := cmd.Run(); err != nil {
		log.ErrorContext(ctx, "Failed to copy to clipboard", "error", err)
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	log.DebugContext(ctx, "Successfully copied to clipboard")
	return nil
}
