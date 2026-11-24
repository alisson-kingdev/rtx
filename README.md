# RTX - Rotacionador de IP (Cross-Platform)

Projeto modular em Go para rotação de IP usando **APIs e proxies públicos e gratuitos por padrão** — sem chaves de API, contas pagas ou proxies proprietários para começar. Suporta **Android (Termux)**, **Windows**, **macOS** e **Linux**.

## Estrutura

```
rtx/
├── cmd/rtx/                # Entry point (Cobra CLI)
│   ├── main.go             # root command, contexto e sinais
│   └── commands/           # subcommands
│       ├── commands.go     # registro e flags compartilhadas
│       ├── rotate.go
│       └── setproxy.go
├── src/
│   ├── controllers/        # Orquestração
│   │   └── rotation.go
│   ├── services/           # Lógica de negócio
│   │   ├── ip.go           # IP público via APIs
│   │   ├── proxy.go        # Coleta de proxies (3 fontes gratuitas)
│   │   ├── rotation.go     # Rotação com limites e fallback
│   │   └── sysproxy.go     # Configuração de proxy do sistema
│   ├── utils/              # Utilitários
│   │   ├── colors.go       # Cores e banner
│   │   ├── network.go      # HTTP com limites + extração de proxies
│   │   ├── validate.go     # Validação de URLs de proxy
│   │   └── shell.go        # Escape seguro para shells
│   └── models/             # Modelos de dados
├── build.sh                # Script de cross-compile
├── Makefile                # Build via Make
├── go.mod
├── go.sum
└── README.md
```

## Fontes gratuitas utilizadas

Proxies públicos (sem autenticação ou chave):

- https://gimmeproxy.com/api/getProxy?curl=true
- https://hproxy.com/api/proxy-list?format=json
- https://api.proxyscraper.com/v1/http.txt

Descoberta de IP público (também gratuitas):

- https://api.ipify.org
- https://icanhazip.com
- https://ifconfig.me
- https://api.ip.sb/ip
- https://ipinfo.io/ip

> Se nenhum proxy público funcionar, o RTX faz fallback para consulta direta às APIs de IP. O comando continua útil sem configurar nada.

## Como usar

### Build para a plataforma atual

```bash
go build -o rtx ./cmd/rtx/
./rtx rotate
./rtx rotate --verbose
./rtx --version
```

### Cross-compile para todas as plataformas

Usando o script:

```bash
chmod +x build.sh
./build.sh all
```

Usando Make:

```bash
make build-all
```

### Validação

```bash
go test ./...
go vet ./...
```

### Plataformas suportadas

| Plataforma | GOOS | GOARCH | Extensão |
|------------|------|--------|----------|
| Linux x64  | linux | amd64 | (nenhuma) |
| Linux ARM64| linux | arm64 | (nenhuma) |
| Linux ARM  | linux | arm | (nenhuma) |
| macOS Intel| darwin | amd64 | (nenhuma) |
| macOS Apple Silicon | darwin | arm64 | (nenhuma) |
| Windows x64| windows | amd64 | .exe |
| Android ARM64 (Termux) | android | arm64 | (nenhuma) |

> Android 32-bit (`android/arm`) exige linking externo com cgo/NDK e não faz parte do `build-all` padrão.

### No Termux (Android)

```bash
pkg install golang
git clone <repo>
cd rtx
./build.sh current
chmod +x rtx
./rtx rotate
```

## Comandos

```bash
rtx rotate                       # Rotação de IP com proxy público ou fallback direto
rtx rotate --verbose             # Modo verboso
rtx rotate --timeout 30s         # Tempo limite total
rtx rotate --max-proxies 10      # Limite de candidatos testados
rtx rotate --concurrency 2       # Testes paralelos

rtx set-proxy default            # Coleta e aplica um proxy público funcional
rtx set-proxy http://host:3128   # Aplica um proxy informado
rtx unset-proxy                  # Remove a configuração gerada pelo RTX

rtx --help                       # Ajuda
rtx --version                    # Versão
```

A flag `--verbose/-v` é global e pode aparecer antes ou depois do subcomando. `Ctrl+C` cancela operações em andamento.

### `rtx set-proxy <URL | default>`

Define `http_proxy`, `https_proxy`, `HTTP_PROXY`, `HTTPS_PROXY`, `no_proxy` e `NO_PROXY` nos arquivos de configuração do shell e, quando aplicável, nas configurações de rede do sistema:

| Plataforma | Arquivos / comandos utilizados |
|------------|-------------------------------|
| **Linux** | `~/.bashrc`, `~/.bash_profile`, `~/.profile`, `~/.shrc` ou `~/.config/fish/config.fish` |
| **macOS** | `networksetup` + arquivos de shell do usuário |
| **Windows** | `setx` (variáveis de usuário) + `netsh winhttp` |
| **Android (Termux)** | `~/.bashrc`, `~/.profile` |

**Comportamento inteligente:**

- `default`coleta proxies das fontes gratuitas, testa cada um e aplica o primeiro que mudar o IP.
- Blocos RTX anteriores são removidos antes de gravar o novo, evitando duplicatas.
- Marcadores: `# >>> RTX PROXY >>>` e `# <<< RTX PROXY <<<`.
- Um bloco sem marcador de fim é preservado — o RTX não apaga configuração de usuário.

Exemplo:

```bash
rtx set-proxy http://proxy.example.com:8080
rtx set-proxy http://user:pass@proxy:3128
rtx set-proxy default
```

### `rtx unset-proxy`

Remove apenas os blocos gerados pelo RTX dos arquivos de configuração do shell e reseta configurações de rede (Windows WinHTTP, macOS networksetup).

## Segurança e robustez

- URLs de proxy são validadas: esquema `http`/`https`, host e porta válidos, sem caminho, query, fragmento ou caracteres de controle.
- Valores gravados no shell são escapados com aspas simples; injeção via URL é rejeitada antes da escrita.
- Credenciais de proxy são exibidas de forma mascarada (`user:xxxxx`) na saída e no resultado do comando.
- Arquivos de configuração são escritos de forma atômica, com permissão `0600`, e symlinks são recusados.
- Respostas HTTP são limitadas a 1 MiB, redirecionamentos cross-host são bloqueados e há timeout por requisição.
- Rotação usa timeout total, limite de candidatos e concorrência limitada.
- `/etc/environment` não é modificado: nenhuma alteração sistêmica silenciosa fora do perfil do usuário.
- Erros são propagados com `error` e contexto; nenhum subcomando chama `os.Exit` internamente.
