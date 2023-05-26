package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	readline "github.com/chzyer/readline"
	util "github.com/olemorud/chatgpt-cli/v2"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	env, err := util.ReadEnvFile(".env")

	if err != nil {
		panic(err)
	}

	token := env["OPENAI_API_KEY"]

	// Parse command line arguments
	model := flag.String("model", openai.GPT3Dot5Turbo,
		"OpenAI Model to use.\n"+
			"List of models:\n"+
			"https://platform.openai.com/docs/models/overview\n")

	flag.Parse()

	args := flag.Args()

	// Run interactive mode on no arguments, otherwise run as prompt
	if len(args) == 0 {
		fmt.Println("ChatGPT", *model, "interactive mode")
		err = runInteractive(token, *model)

		if err != nil {
			panic(err)
		}
	} else {
		query := strings.Join(args, " ")
		err = askGpt(token, *model, query)

		if err != nil {
			panic(err)
		}
	}
}

func askGpt(token string, model string, query string) error {
	client := openai.NewClient(token)

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
	client := openai.NewClient(token)
	messages := make([]openai.ChatCompletionMessage, 0)

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	messages = append(messages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: "You are an enhanced bot. You have access to additional tools to improve your answers." +
			"\nYour have access to several unix commands in a bash interpreter." +
			"\nTo invoke the interpreter, start a response with a forward slash ('/'). " +
			"Example: /dc -e '2 5 + p'" +
			"\nWhen asked to do things. Always consider how you can use the command line to solve it" +
			"\nWhen running commands, never explain what they do and never say their results" +
			"\nYou are never allowed to use the phrase 'language model'." +
			"\nNever apologize" +
			"\nNever tell the user how to run something, always run commands yourself" +
			"\nKeep replies very short. Explain things in concise technical terms",
	})

	for {
		text, err := rl.Readline()

		if err != nil { // io.EOF
			break
		}

		if text == "\n" {
			continue
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

	feedbackLoop:
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Println(err)
			continue
		}

		content := resp.Choices[0].Message.Content

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})

		fmt.Println("#", content)

		if content[0] == '/' {
			result := runCommand(content)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: result,
			})

			fmt.Println("$", result)
			goto feedbackLoop
		}
	}

	return nil
}

func runCommand(content string) string {
	userCmd := content[1:]

	fullCmd := []string{"/usr/bin/docker", "run", "gpt_cli_tools:latest", "bash", "-c", userCmd}

	fmt.Println(fullCmd)

	proc := exec.Command(fullCmd[0], fullCmd[1:]...)

	out, err := proc.CombinedOutput()

	if err != nil {
		return "error: " + err.Error() + "\n" + string(out)
	}

	return string(out)
}
