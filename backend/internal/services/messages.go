package services

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tmc/langchaingo/llms"
)

type CreateSystemMessageArgs struct {
	ChatID  string
	Content string
}

func CreateSystemMessage(args CreateSystemMessageArgs, Q *db.Queries, ctx context.Context) error {
	if Q == nil {
		Q = shared.Q
	}

	if ctx == nil {
		ctx = context.Background()
	}

	id, _ := gonanoid.New()

	if err := Q.InsertMessage(ctx, db.InsertMessageParams{
		ID:   id,
		Role: string(llms.ChatMessageTypeSystem),
		Content: sql.NullString{
			String: args.Content,
			Valid:  args.Content != "",
		},
		StopReason: sql.NullString{Valid: false},
		ChatID:     args.ChatID,
	}); err != nil {
		return fmt.Errorf("Failed to create system message - %w\n", err)
	}

	return nil
}

type CreateHumanMessageArgs struct {
	ChatID  string
	Content string
}

func CreateHumanMessage(args CreateHumanMessageArgs, Q *db.Queries, ctx context.Context) error {
	if Q == nil {
		Q = shared.Q
	}

	if ctx == nil {
		ctx = context.Background()
	}

	id, _ := gonanoid.New()

	if err := Q.InsertMessage(ctx, db.InsertMessageParams{
		ID:         id,
		Role:       string(llms.ChatMessageTypeHuman),
		Content:    sql.NullString{String: args.Content, Valid: args.Content != ""},
		StopReason: sql.NullString{Valid: false},
		ChatID:     args.ChatID,
	}); err != nil {
		return fmt.Errorf("Failed to create human message - %w\n", err)
	}

	return nil
}

type CreateAIMessageArgs struct {
	ChatID     string
	StopReason string
}

func CreateAIMessage(args CreateAIMessageArgs, Q *db.Queries, ctx context.Context) (AIMessageID string, err error) {
	if Q == nil {
		Q = shared.Q
	}

	if ctx == nil {
		ctx = context.Background()
	}

	id, _ := gonanoid.New()

	if err := Q.InsertMessage(ctx, db.InsertMessageParams{
		ID:         id,
		Role:       string(llms.ChatMessageTypeAI),
		Content:    sql.NullString{Valid: false},
		StopReason: sql.NullString{String: args.StopReason, Valid: args.StopReason != ""},
		ChatID:     args.ChatID,
	}); err != nil {
		return "", fmt.Errorf("Failed to create AI message - %w\n", err)
	}

	return id, nil
}

type InsertAITextPartArgs struct {
	AIMessageID string
	Content     string
	PartIndex   int64
}

func InsertAITextPart(args InsertAITextPartArgs, Q *db.Queries, ctx context.Context) (err error) {
	var tx *sql.Tx

	if Q == nil {
		tx, err := shared.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("Failed to begin tx - %w", err)
		}
		defer tx.Rollback()

		Q = shared.Q.WithTx(tx)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	partID, err := Q.InsertAIMessagePart(ctx, db.InsertAIMessagePartParams{
		Type:      "text",
		PartIndex: args.PartIndex,
		MessageID: args.AIMessageID,
	})

	if err != nil {
		return fmt.Errorf("Failed to insert AI message part - %w", err)
	}

	if err := Q.InsertTextPart(ctx, db.InsertTextPartParams{
		Text: sql.NullString{
			String: args.Content,
			Valid:  args.Content != "",
		},
		MessagePartID: partID,
	}); err != nil {
		return fmt.Errorf("Failed to insert text part - %w", err)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("Failed to commit text part to DB - %w", err)
		}
	}

	return nil
}

type InsertAIToolCallRequestPartArgs struct {
	AIMessageID string
	ToolCallID  string
	ToolName    string
	Arguments   string
	PartIndex   int64
}

func InsertAIToolCallRequestPart(args InsertAIToolCallRequestPartArgs, Q *db.Queries, ctx context.Context) (err error) {
	var tx *sql.Tx

	if Q == nil {
		tx, err = shared.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("Failed to begin tx - %w", err)
		}
		defer tx.Rollback()

		Q = shared.Q.WithTx(tx)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	fmt.Println("Creating AI Message")

	partID, err := Q.InsertAIMessagePart(ctx, db.InsertAIMessagePartParams{
		Type:      "function",
		PartIndex: args.PartIndex,
		MessageID: args.AIMessageID,
	})

	if err != nil {
		return fmt.Errorf("Failed to insert AI message - %w", err)
	}

	fmt.Println("Creating Tool Call Part")

	if err := Q.InsertToolCallPart(ctx, db.InsertToolCallPartParams{
		ToolCallID:    args.ToolCallID,
		Name:          args.ToolName,
		Arguments:     args.Arguments,
		MessagePartID: partID,
	}); err != nil {
		return fmt.Errorf("Failed to insert tool call req part - %w", err)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("Failed to commit tool call req part to DB - %w", err)
		}
	}

	return nil
}

type InsertToolCallResultArgs struct {
	ChatID        string
	ToolCallID    string
	ToolName      string
	ResultContent string
	IsError       bool
}

func InsertToolCallResult(args InsertToolCallResultArgs, Q *db.Queries, ctx context.Context) (err error) {
	var tx *sql.Tx

	if Q == nil {
		tx, err = shared.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("Failed to begin tx - %w", err)
		}
		defer tx.Rollback()

		Q = shared.Q.WithTx(tx)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	id, _ := gonanoid.New()
	if err := Q.InsertMessage(ctx, db.InsertMessageParams{
		ID:         id,
		Role:       string(llms.ChatMessageTypeTool),
		Content:    sql.NullString{Valid: false},
		StopReason: sql.NullString{Valid: false},
		ChatID:     args.ChatID,
	}); err != nil {
		return fmt.Errorf("Failed to insert tool call result message - %w", err)
	}
	
	if err := Q.InsertToolCallResult(ctx, db.InsertToolCallResultParams{
		MessageID:  id,
		ToolCallID: args.ToolCallID,
		Name:       args.ToolName,
		Content: func() string {
			if args.IsError {
				return "ERROR: " + args.ResultContent
			}

			return args.ResultContent
		}(),
		IsError: args.IsError,
	}); err != nil {
		return fmt.Errorf("Failed to insert tool call result - %w", err)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("Failed to commit tool call result to DB - %w", err)
		}
	}

	return nil
}
