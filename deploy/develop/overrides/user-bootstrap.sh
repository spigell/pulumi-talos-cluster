#!/usr/bin/env bash
set -euo pipefail

USER_NAME="spigell"
USER_UID=7000
USER_HOME="/home/${USER_NAME}"

ensure_user() {
  if ! id -u "${USER_NAME}" >/dev/null 2>&1; then
    echo ">>> Creating user ${USER_NAME} (${USER_UID})"
    useradd -m -u "${USER_UID}" -s /bin/bash "${USER_NAME}"
  fi
  
  cat >> "/etc/profile" <<'EOF'
export PATH=/pulumi/bin:/usr/local/share/fnm/aliases/default/bin:/usr/local/share/pyenv/shims:/usr/local/share/pyenv/bin:/usr/local/share/dotnet:~/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.pulumi/bin
export PULUMI_CONFIG_PASSPHRASE=1234
EOF

  mkdir -p "${USER_HOME}"/{.home,.cache/gomod,.cache/gobuild,.cache/gopath,go/bin,pulumi-talos-cluster}
  chown -R "${USER_UID}:${USER_UID}" "${USER_HOME}"

  cat >> "${USER_HOME}/.bashrc" <<'EOF'
# Completions (ignore errors if not present)
[ -f /usr/share/bash-completion/completions/make ] && source /usr/share/bash-completion/completions/make
command -v pulumi >/dev/null 2>&1 && source <(pulumi gen-completion bash) || true
command -v talosctl >/dev/null 2>&1 && source <(talosctl completion bash) || true

pdebug() {
  local log="/tmp/pulumi-provider.log"
  if [[ ! -s "$log" ]]; then
    echo "pdebug: provider log '$log' not found or empty" >&2
    return 1
  fi

  # last non-empty line
  local port
  port="$(awk 'NF{last=$0} END{print last}' "$log")"

  if [[ ! "$port" =~ ^[0-9]+$ ]]; then
    echo "pdebug: expected a bare port number on the last log line, got: '$port'" >&2
    return 1
  fi

  echo "pdebug: using talos-cluster provider port $port"

  if [[ -n "${DEBUG_COMMAND:-}" ]]; then
    PULUMI_DEBUG_PROVIDERS="talos-cluster:${port}" \
      pulumi ${DEBUG_COMMAND} --logtostderr -d -v 12 "$@" 2> /tmp/log.txt
  else
    PULUMI_DEBUG_PROVIDERS="talos-cluster:${port}" \
      pulumi "$@" --logtostderr -d -v 12 2> /tmp/log.txt
  fi
}
alias pulumi-dbg='pdebug'
EOF
  chown "${USER_UID}:${USER_UID}" "${USER_HOME}/.bashrc"
}

run_as_user() { # $@ = command
  su - "${USER_NAME}" -s /bin/bash -c "$*"
}


case "${1:-}" in
  init)
    ensure_user
    echo ">>> Pulumi login (local) for ${USER_NAME}"
    run_as_user "pulumi login --local || true"
    echo ">>> Build plugin & start delve"
    run_as_user "cd /projects/pulumi-talos-cluster && make start_delve | tee > /tmp/pulumi-provider.log"
    ;;
  main)
    ensure_user
    echo ">>> Pulumi login (local) for ${USER_NAME}"
    run_as_user "pulumi login --local || true"
    echo '>>> Ready; keeping container alive'
    run_as_user "sleep infinity"
    ;;
  *)
    echo "Usage: $0 {init|main}" >&2
    exit 1
    ;;
esac