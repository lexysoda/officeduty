package controller

import (
	"fmt"
	"github.com/lexysoda/officeduty/rotation"
	"github.com/slack-go/slack"
	"log/slog"
)

func (c *Controller) sendTextMessage(text string) {
	c.s.PostMessage(c.channel, slack.MsgOptionText(text, false))
}

func (c *Controller) sendBlockMessage(blocks []slack.Block) {
	c.s.PostMessage(c.channel, slack.MsgOptionBlocks(blocks...))
}

func (c *Controller) SendEphemeralBlockMessage(u rotation.User, blocks []slack.Block) {
	c.s.PostEphemeral(c.channel, string(u), slack.MsgOptionBlocks(blocks...))
}

func (c *Controller) sendUserShift(u rotation.User) {
	blocks, err := c.userShiftBlocks(u)
	if err != nil {
		return
	}
	c.s.PostEphemeral(
		c.channel,
		string(u),
		slack.MsgOptionBlocks(blocks...),
	)
}

func (c *Controller) sendUserShiftReplace(u rotation.User, responseURL string) {
	//blocks, err := c.userShiftBlocks(u)
	//if err != nil {
	//	return
	//}
	c.s.PostEphemeral(
		c.channel,
		string(u),
		slack.MsgOptionBlocks([]slack.Block{
			slack.NewHeaderBlock(
				&slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "great success",
				},
			),
		}...),
		slack.MsgOptionReplaceOriginal(responseURL),
	)
}

func (c *Controller) userShiftBlocks(user rotation.User) ([]slack.Block, error) {
	date, err := c.r.NextUserShift(user)
	if err != nil {
		return nil, err
	}
	start := date.Format("Jan 02")
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Your Next Shift",
			},
		),
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: "*Start:* " + start,
			},
			nil,
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: "Push back your shift to the next cycle:",
			},
			nil,
			slack.NewAccessory(
				slack.NewButtonBlockElement(
					"reschedule",
					string(user),
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Reschedule",
					},
				).WithStyle(slack.StyleDanger),
			),
		),
	}
	return blocks, nil
}

func (c *Controller) sendShiftPlan() {
	s := c.r.NextShift()
	start := s.Start.Format("Jan 02")
	end := s.End.Format("Jan 02")
	users := s.Users
	slog.Debug("Sending shift plan", "start", start, "end", end, "users", users)

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(
				"plain_text",
				":rotating_light:   New Office Duty Shift   :rotating_light:",
				false,
				false,
			),
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				slack.NewTextBlockObject(
					"mrkdwn",
					fmt.Sprintf("*Start:* %s", start),
					false,
					false,
				),
				slack.NewTextBlockObject(
					"mrkdwn",
					fmt.Sprintf("*End:* %s", end),
					false,
					false,
				),
			},
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewRichTextBlock(
			"",
			slack.NewRichTextList(
				slack.RTEListBullet,
				0,
				slack.NewRichTextSection(
					users[0].ToBlockElement(),
				),
				slack.NewRichTextSection(
					users[1].ToBlockElement(),
				),
				slack.NewRichTextSection(
					users[2].ToBlockElement(),
				),
			),
		),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject(
				"mrkdwn",
				"<https://stroeerdigitalgroup.atlassian.net/wiki/spaces/MBR/pages/3528261726/Labs+Berlin+Permanent+Office+Management+and+Maintenance|Click here for a list of your responsibilities.>",
				false,
				false,
			),
			nil,
			nil,
		),
		slack.NewDividerBlock(),
		slack.NewContextBlock(
			"",
			slack.NewTextBlockObject(
				"mrkdwn",
				"Use `/officeshift` to see your next shift and the option to reschedule.",
				false,
				false,
			),
			slack.NewTextBlockObject(
				"mrkdwn",
				"Use `@Officeduty add ${user}` to add ${user} to the rotation.",
				false,
				false,
			),
			slack.NewTextBlockObject(
				"mrkdwn",
				"Use `@Officeduty` to send the current shift (this message).",
				false,
				false,
			),
		),
	}

	c.sendBlockMessage(blocks)
}

func (c *Controller) sendFailedReschedule(u rotation.User, responseURL string) {
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Failed to reschedule",
			},
		),
	}
	c.s.PostEphemeral(
		c.channel,
		string(u),
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionReplaceOriginal(responseURL),
	)
}
