package system

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/asdf57/bsw/cmd/bot/shared"
	"github.com/bwmarrin/discordgo"
)

func HandleExportData(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := shared.RespondDeferred(s, i); err != nil {
		return err
	}

	data, err := shared.ExportCheckpoint()
	if err != nil {
		_, followupErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: fmt.Sprintf("Export failed: %s", err.Error())})
		return followupErr
	}

	filename := fmt.Sprintf("bsw-checkpoint-%s.json", time.Now().UTC().Format("20060102-150405"))
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Export ready.",
		Files: []*discordgo.File{{
			Name:        filename,
			ContentType: "application/json",
			Reader:      bytes.NewReader(data),
		}},
	})
	return err
}

func HandleImportData(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := shared.RespondDeferred(s, i); err != nil {
		return err
	}

	attachment, err := checkpointAttachment(i.ApplicationCommandData())
	if err != nil {
		_, followupErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: err.Error()})
		return followupErr
	}

	data, err := downloadAttachment(attachment)
	if err != nil {
		_, followupErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: fmt.Sprintf("Import failed: %s", err.Error())})
		return followupErr
	}

	result, err := shared.ImportCheckpoint(data)
	if err != nil {
		_, followupErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: fmt.Sprintf("Import failed: %s", err.Error())})
		return followupErr
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Import complete: %d users, %d tags, %d payments, %d debts, %d settlements, %d exchange rates.",
			result.Users,
			result.Tags,
			result.Payments,
			result.Debts,
			result.Settlements,
			result.ExchangeRates,
		),
	})
	return err
}

func checkpointAttachment(data discordgo.ApplicationCommandInteractionData) (*discordgo.MessageAttachment, error) {
	option := data.GetOption("file")
	if option == nil {
		return nil, fmt.Errorf("Attach a checkpoint JSON file.")
	}
	if data.Resolved == nil {
		return nil, fmt.Errorf("Could not resolve the uploaded file.")
	}

	attachmentID, ok := option.Value.(string)
	if !ok || strings.TrimSpace(attachmentID) == "" {
		return nil, fmt.Errorf("Could not read the uploaded file ID.")
	}

	attachment, ok := data.Resolved.Attachments[attachmentID]
	if !ok || attachment == nil {
		return nil, fmt.Errorf("Could not find the uploaded file.")
	}
	if attachment.Size > 8*1024*1024 {
		return nil, fmt.Errorf("Checkpoint file is too large.")
	}
	if !strings.HasSuffix(strings.ToLower(attachment.Filename), ".json") {
		return nil, fmt.Errorf("Checkpoint file must be JSON.")
	}
	return attachment, nil
}

func downloadAttachment(attachment *discordgo.MessageAttachment) ([]byte, error) {
	resp, err := http.Get(attachment.URL)
	if err != nil {
		return nil, fmt.Errorf("download attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024+1))
	if err != nil {
		return nil, fmt.Errorf("read attachment: %w", err)
	}
	if len(data) > 8*1024*1024 {
		return nil, fmt.Errorf("checkpoint file is too large")
	}
	return data, nil
}
