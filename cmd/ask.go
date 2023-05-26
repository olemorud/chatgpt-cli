package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	util "github.com/olemorud/chatgpt-cli/v2"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	env, err := util.ReadEnvFile(".env")

	if err != nil {
		fmt.Println("failed to read .env", err)
	}

	// parse command line arguments
	token := env["OPENAI_API_KEY"]

	model := *flag.String("model", openai.GPT3Dot5Turbo,
		"OpenAI Model to use.\n"+
			"List of models:\n"+
			"https://platform.openai.com/docs/models/overview\n")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		runInteractive(token, model)
	} else {
		query := strings.Join(args, " ")
		err = askGpt(token, model, query)

		if err != nil {
			panic(err)
		}
	}
}

func askGpt(token string, model string, query string) error {
	client := openai.NewClient("your token")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(resp.Choices[0].Message.Content)

	return nil
}

func runInteractive(token string, model string) error {
	client := openai.NewClient("your token")
	messages := make([]openai.ChatCompletionMessage, 0)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("ChatGPT", model, "interactive mode")

	for scanner.Scan() {
		fmt.Println("->")

		text := scanner.Text()

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
			},
		)

		fmt.Println(resp.Choices[0].Message.Content)

		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return nil
}
