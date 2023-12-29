# README for Chat-CLI: An AI-Powered CLI Tool in Go

## Overview

Chat-CLI is an innovative command-line interface (CLI) tool, powered by AI, specifically designed to facilitate users in the Unix Terminal environment. This tool harnesses the capabilities of AI to provide expert guidance and assistance in a range of Unix Terminal tasks. It's ideal for both beginners and advanced users, offering support in command syntax, scripting, troubleshooting, system management, and optimizing workflows.

## Features

- **AI-Powered Assistance**: Utilizes OpenAI's GPT-4 model to offer real-time, context-aware support.
- **Flexible Configuration**: Supports custom configurations for API keys and assistant settings.
- **Dynamic Assistant Management**: Allows users to create, interact with, update, and delete AI assistants.
- **User-Friendly Interface**: Designed to enhance the Unix Terminal experience with intuitive interactions.
- **Robust Error Handling**: Implements comprehensive error checking and informative messaging.

## Installation

_Note: Chat-CLI requires Go to be installed on your system._

1. Clone the repository from [GitHub link].
2. Navigate to the cloned directory.
3. Run `go build` to compile the source code.

## Configuration

Before using Chat-CLI, you must configure it with your OpenAI API key. You can do this in by using the tool's prompts to input your API key when you first run it. Your key _never_ leaves your machine!

## Usage

To start Chat-CLI, run the compiled executable. The tool offers the following options:

- Create a new assistant
- Interact with an existing assistant

## Development TODOs

- Implement a simple menu for selecting an existing conversation.
- Enhance logging with structured logging libraries.
- Add context handling for cancellation and timeouts in HTTP requests.
- Write unit tests for each method.
- Introduce an advanced configuration system.
- List, update, or delete existing assistants
- Retrieve details about assistants, conversations, messages, tools, or users
- Implement an improved CLI using libraries like `tview`, `cobra`, or `urfave/cli`.
