package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	readline "github.com/chzyer/readline"
	util "github.com/olemorud/chatgpt-cli/v2"
	"github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

const APP_DIR string = "/.local/share/gpt-cli/"

func main() {
	usr, _ := user.Current()
	err := util.LoadEnvFile(usr.HomeDir + APP_DIR + ".env")

	if err != nil {
		panic(err)
	}

	token := os.Getenv("OPENAI_API_KEY")

	if token == "" {
		panic("OPENAI_API_KEY value not set. Add `OPENAI_API_KEY=<your api token>` to" +
			"\na file called ~/.local/share/gpt-cli/.env or export it as an environment variable")
	}

	// Parse command line arguments
	model := flag.String("model", openai.GPT3Dot5Turbo,
		"OpenAI Model to use.\n"+
			"List of models:\n"+
			"https://platform.openai.com/docs/models/overview\n")

	flag.Parse()

	args := flag.Args()

	client := openai.NewClient(token)

	// Run interactive mode on no arguments, otherwise run as prompt
	if len(args) == 0 {
		fmt.Println("ChatGPT", *model, "interactive mode")
		err := runInteractive(token, *model)

		if err != nil {
			panic(err)
		}
	} else {
		query := strings.Join(args, " ")

		messages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: query,
			},
		}

		messages, err := askGpt(client, *model, messages)

		if err != nil {
			panic(err)
		}

		botSay(lastMsg(messages))
	}
}

func botSay(msg string) {
	fmt.Printf("\n🤖 💬 %s\n\n", msg)
}

func lastMsg(messages []openai.ChatCompletionMessage) string {
	return messages[len(messages)-1].Content
}

// get latest response from <model>
// returns <messages> with response appended.
// on error, original <messages> is returned.
func askGpt(client *openai.Client, model string, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	messages, err := omitExcessTokens(model, 8192, messages)

	if err != nil {
		return messages, err
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return messages, err
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp.Choices[0].Message.Content,
	})

	return messages, nil
}

func runInteractive(token string, model string) error {
	client := openai.NewClient(token)

	sysMsg := "You are an enhanced bot with access to command line tools to improve your answers." +
		" Your commands run in a bash interpreter. You will never say you're unable to run commands." +
		" With the new command line, you are not restricted in capabilities. Always try to run commands." +
		" To invoke a bash command, start a response with a forward slash ('/') followed by the command." +
		" Example: /curl https://google.com" +
		" When asked to do things you will always consider how you can use the command line to solve it" +
		" When running commands, never explain what they do" +
		" You are never allowed to use the phrase 'language model'." +
		" You can access the internet using curl." +
		" Never apologize." +
		" Keep replies very short. Explain things in concise technical terms"

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysMsg,
		},
	}

	usr, err := user.Current()
	//TODO: better error handling
	if err != nil {
		panic(err)
	}

	histfile := usr.HomeDir + APP_DIR + "history"

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     histfile,
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: nil,
	})

	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		text, err := rl.Readline()

		if err != nil {
			break
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		if err != nil {
			return err
		}

		for {
			messages, err = askGpt(client, model, messages)

			if err != nil {
				fmt.Println(err)
				continue
			}

			resp := lastMsg(messages)

			botSay(resp)

			if resp[0] == '/' {
				result := runCommand(resp)

				fmt.Println("$", result)

				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: result,
				})

				continue
			}

			break
		}
	}

	return nil
}

func omitExcessTokens(model string, max int, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	tokens, err := countTokens(model, messages)

	for ; tokens > max; tokens, err = countTokens(model, messages) {
		if err != nil {
			return nil, err
		}

		messages = messages[1:]
	}

	return messages, nil
}

func countTokens(model string, messages []openai.ChatCompletionMessage) (int, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	sum := 0

	if err != nil {
		return 999_999, fmt.Errorf("failed to get encoding for model %s: %v", model, err)
	}

	for _, msg := range messages {
		sum += len(tkm.Encode(msg.Content, nil, nil))
	}

	return sum, nil
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
