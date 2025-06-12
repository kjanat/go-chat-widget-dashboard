# Use the official Go devcontainer image as base
FROM mcr.microsoft.com/devcontainers/go:1-bookworm

# Environment variables
ENV NODE_VERSION=22
ENV SHELL=/bin/zsh
ENV ZSH_THEME="robbyrussell"

# Install system dependencies and tools in one layer
RUN sudo apt-get update \
  && sudo apt-get upgrade -y \
  && sudo apt-get autoremove -y \
  && sudo apt-get clean

RUN sudo apt-get install -y \
  apt-transport-https

RUN sudo apt-get install -y \
  ca-certificates

RUN sudo apt-get install -y \
  curl

RUN sudo apt-get install -y \
  gnupg

RUN sudo apt-get install -y \
  lsb-release

RUN sudo apt-get install -y \
  software-properties-common

RUN sudo apt-get install -y \
  zsh

RUN sudo apt-get install -y \
  zsh-autosuggestions

RUN sudo apt-get install -y \
  zsh-syntax-highlighting

RUN sudo apt-get install -y \
  zsh-theme-powerlevel9k

RUN sudo apt-get install -y \
  tmux

RUN sudo apt-get install -y \
  ripgrep

RUN sudo apt-get install -y \
  shellcheck

RUN sudo apt-get install -y \
  sqlite3

RUN sudo apt-get install -y \
  upx-ucl

RUN sudo apt-get install -y \
  python3-pip

RUN sudo apt-get install -y \
  wget

# Install Node.js directly instead of using NVM to avoid shell sourcing issues
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
  && apt-get install -y nodejs

# Install zsh and oh-my-zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
  && git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions

# Install pnpm and global npm packages
RUN npm install -g pnpm @devcontainers/cli how2 ts-node tsx @vue/cli prettier typescript \
  && printf "alias pn=pnpm\nalias npm=pnpm\n" >> ~/.zshrc

# Install Docker in Docker
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
  && apt-get update \
  && apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
  && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
  && apt-get update \
  && apt-get install -y gh

# Install Go tools
RUN go install github.com/arl/gitmux@latest \
  && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
  && go install github.com/a-h/templ/cmd/templ@latest

# Install PowerShell
RUN wget -q https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb \
  && dpkg -i packages-microsoft-prod.deb \
  && rm packages-microsoft-prod.deb \
  && apt-get update \
  && apt-get install -y powershell

# Install additional tools (thefuck, golangci-lint, mcfly, fzf)
RUN pip3 install thefuck \
  && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin latest \
  && curl -LSfs https://raw.githubusercontent.com/cantino/mcfly/master/ci/install.sh | sh -s -- --git cantino/mcfly \
  && git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf \
  && ~/.fzf/install --all \
  && echo 'eval "$(mcfly init zsh)"' >> ~/.zshrc

# Set zsh as default shell
RUN chsh -s $(which zsh) root

# Configure zsh plugins
RUN sed -i 's/plugins=(git)/plugins=(ssh-agent npm git colored-man-pages colorize common-aliases command-not-found copypath debian docker docker-compose encode64 extract frontend-search fzf gh gitignore golang gpg-agent isodate magic-enter npm nvm pnpm ssh systemd thefuck timer tmux vscode urltools zsh-autosuggestions)/' ~/.zshrc

# Clean up
RUN apt-get autoremove -y \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /workspace
