# Prompt Maker

**A CLI tool to help you craft the perfect prompt for Gemini models using a two-step refinement process.**

`prompt-maker` is an interactive terminal application built with Go and Bubble Tea. It takes your rough idea for a prompt, sends it to a Gemini model with a specialized "prompt optimization" system prompt, and returns a beautifully crafted, detailed prompt. You can then review this new prompt and resubmit it to get your final, high-quality answer.

## Features

*   **Interactive Model Selection**: Choose from the latest Gemini models at startup.
*   **Two-Step Prompt Refinement**:
    1.  Provide a rough prompt.
    2.  Receive a detailed, optimized prompt crafted by the AI.
    3.  Resubmit the optimized prompt to get your final answer.
*   **Polished Terminal UI**: A clean, full-screen interface built with the Bubble Tea framework.
*   **Keyboard Shortcuts**: Intuitive shortcuts for copying (`c`), resubmitting (`r`), and quitting (`esc`).
*   **Dynamic Versioning**: The application version is injected at build time for easy tracking.

## Installation

### Prerequisites

*   Go (version 1.18 or newer is recommended).
*   Git.

### Building from Source

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/chunghha/prompt-maker.git
    cd prompt-maker
    ```

2.  **Build the application:**
    This project uses `Task` for simple build commands. If you have `Task` installed, you can build the project with:
    ```bash
    task build
    ```    This will create a binary named `prompt_maker` in the project root.

    Alternatively, you can use the standard Go build command:
    ```bash
    go build -ldflags="-X 'prompt-maker/cmd.version=1.0.1'" -o prompt_maker .
    ```

3.  **Install (Optional):**
    You can move the `prompt_maker` binary to a directory in your system's `PATH` for easy access.
    ```bash
    mv prompt_maker /usr/local/bin/
    ```

## Usage

### 1. Configuration

Before running the application, you must set your Google AI API key as an environment variable:

```bash
export GEMINI_API_KEY="your_google_ai_api_key"
```

### 2. Running the Application

Execute the binary from your terminal:

```bash
./prompt_maker
```

### 3. The Workflow

The application follows a simple two-step process:

1.  **Select a Model**: Use the arrow keys to choose a Gemini model and press `Enter`.
2.  **Enter a Rough Prompt**: Type your basic idea (e.g., "an email to my boss asking for a raise") and press `Enter`.
3.  **Review the Crafted Prompt**: The application will display a detailed, optimized prompt.
4.  **Resubmit or Edit**:
    *   Press `r` to immediately resubmit the crafted prompt to get your final answer.
    *   Alternatively, you can type a new prompt.
5.  **Get the Final Answer**: The final response from the model will be displayed.
6.  **Copy or Quit**:
    *   Press `c` to copy the response to your clipboard.
    *   Press `Enter` to start over or `esc` to quit.

### Keyboard Shortcuts

| Key     | Action                                     | Context                               |
| :------ | :----------------------------------------- | :------------------------------------ |
| `Enter` | Submit prompt                              | When entering a rough prompt          |
| `r`     | **R**esubmit the crafted prompt            | After a prompt has been crafted       |
| `c`     | **C**opy the response to the clipboard     | After a prompt or answer is displayed |
| `esc`   | Quit the application                       | At any time                           |

## Development

This project uses `Taskfile.yml` to manage common development tasks.

*   **Build the binary**: `task build`
*   **Run all tests**: `task test`
*   **Run linters**: `task lint`
*   **Format the code**: `task format`
*   **Install the binary**: `task install` (copies it to `~/bin/`)

___

### Credit

The core "Lyra" prompt optimization methodology was inspired by a post from Min Choi. [View the original post on X](https://x.com/minchoi/status/1940251597050646766).
