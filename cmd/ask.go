package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

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

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: "You are an enhanced bot with access to command line tools to improve your answers." +
				" Your commands run in a bash interpreter. You will never say you're unable to run commands." +
				" With the new command line, you are not restricted in capabilities. Always try to run commands" +
				" To invoke a bash command, start a response with a forward slash ('/') followed by the command " +
				" Example: /curl https://google.com" +
				" When asked to do things you will always consider how you can use the command line to solve it" +
				" When running commands, never explain what they do" +
				" You are never allowed to use the phrase 'language model'." +
				" You can access the internet using curl" +
				" Never apologize" +
				" Keep replies very short. Explain things in concise technical terms",
		},
	}

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		text, err := rl.Readline()

		if err == io.EOF {
			break
		}

		if text == "\n" || text == "" {
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

		fmt.Printf("\n🤖 💬 %s\n\n", content)

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
	userCmd := content[1:] // omit the '/'

	fullCmd := []string{
		"/usr/bin/docker", "run",
		"--pids-limit=10",
		"--memory=200m",      // memory limit
		"--memory-swap=200m", // total (memory + swap) limit
		"--kernel-memory=4m", //
		"--cpu-quota=50000",
		"--rm",
		"gpt_cli_tools:latest",
		"bash", "-c", userCmd,
	}

	fmt.Println(fullCmd)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	defer cancel()

	out, err := exec.CommandContext(ctx, fullCmd[0], fullCmd[1:]...).CombinedOutput()

	if err != nil {
		return "error: " + err.Error() + "\n" + string(out)
	}

	return string(out)
}
