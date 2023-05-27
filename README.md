
# ChatGPT command line interface

ChatGPT in the command line, extended with the ability to run bash commands
(in a container). Although the host file system is not exposed it can still
hypothetically fork bomb your computer. Use at your own risk.

## Build

```sh
go build cmd/ask.go
```

## Example usage

### Run single query

```sh
 $ ./ask write a python array of 10 nouns

nouns = ["apple", "cat", "book", "car", "tree", "house", "pen", "computer", "phone", "water"]
```

### Run (Interactive mode)

With gpt-3.5-turbo (default)
```sh
$ ./ask
ChatGPT gpt-3.5-turbo interactive mode
>
```

With gpt-4
```sh
./ask --model=gpt-4
```

gpt-4 will usually utilize its command line to solve tasks that it traditionally
struggles with, such as maths or counting words:
```sh
./ask --model=gpt-4
ChatGPT gpt-4 interactive mode
> how many words are in the sentence "hello dear world!"?
# /echo "hello dear world!" | wc -w
$ 3

# Yes, there are 3 words in the sentence "hello dear world!"
```