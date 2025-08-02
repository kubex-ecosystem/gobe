# GoBE - Modular & Secure Back-end

![GoBE Banner](/docs/assets/top_banner_lg_b.png)

[![Build Status](https://img.shields.io/github/actions/workflow/status/rafa-mori/gobe/release.yml?branch=main)](https://github.com/rafa-mori/gobe/actions)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/rafa-mori/gobe/blob/main/LICENSE)
[![Automation](https://img.shields.io/badge/automation-zero%20config-blue)](#features)
[![Modular](https://img.shields.io/badge/modular-yes-yellow)](#features)
[![Security](https://img.shields.io/badge/security-high-red)](#features)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/rafa-mori/gobe/blob/main/CONTRIBUTING.md)

---

## **Table of Contents**

1. [About the Project](#about-the-project)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
    - [CLI](#cli)
    - [Configuration](#configuration)
5. [Roadmap](#roadmap)
6. [Contributing](#contributing)
7. [Contact](#contact)

---

## **About the Project**

GoBE é um back-end modular desenvolvido em Go, focado em **segurança, automação e flexibilidade**. Pode rodar como **servidor principal** ou ser utilizado **como módulo** para gerenciamento de funcionalidades como **criptografia, certificados, middlewares, logging e autenticação**.

- **Zero-config:** Não exige configuração manual, gera todos os certificados e armazena informações sensíveis de forma segura no keyring do sistema.
- **Extensível:** Pode ser integrado a outros sistemas ou rodar standalone.

---

## **Features**

✨ **Totalmente modular**

- Todas as lógicas seguem interfaces bem definidas, garantindo encapsulamento.
- Pode ser usado como servidor ou como biblioteca/módulo.

🔒 **Zero-config, mas personalizável**

- Roda sem configuração inicial, mas aceita customização via arquivos.
- Gera certificados, senhas e configurações seguras automaticamente.

🔗 **Integração direta com `gdbase`**

- Gerenciamento de bancos de dados via Docker.
- Otimizações automáticas para persistência e performance.

🛡️ **Autenticação avançada**

- Certificados gerados dinamicamente.
- Senhas aleatórias e keyring seguro.

🌐 **API REST robusta**

- Endpoints para autenticação, gerenciamento de usuários, produtos, clientes e cronjobs.

📋 **Gerenciamento de logs e segurança**

- Rotas protegidas, armazenamento seguro e monitoramento de requisições.

🧑‍💻 **CLI poderosa**

- Comandos para iniciar, configurar e monitorar o servidor.

---

## **Installation**

Requisitos:

- Go 1.19+
- Docker (para integração com bancos via gdbase)

Clone o repositório e compile o GoBE:

```sh
# Clone o repositório
git clone https://github.com/rafa-mori/gobe.git
cd gobe
go build -o gobe .
```

---

## **Usage**

### CLI

Inicie o servidor principal:

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

Isso inicializa o servidor, gera certificados, configura bancos de dados e começa a escutar requisições!

Veja todos os comandos disponíveis:

```sh
./gobe --help
```

**Principais comandos:**

| Comando   | Função                                             |
|-----------|----------------------------------------------------|
| `start`   | Inicializa o servidor                              |
| `stop`    | Encerra o servidor de forma segura                 |
| `restart` | Reinicia todos os serviços                         |
| `status`  | Exibe o status do servidor e dos serviços ativos   |
| `config`  | Gera um arquivo de configuração inicial            |
| `logs`    | Exibe os logs do servidor                          |

---

### Configuration

O GoBE pode rodar sem configuração inicial, mas aceita customização via arquivos YAML/JSON. Por padrão, tudo é gerado automaticamente no primeiro uso.

Exemplo de configuração:

```yaml
port: 3666
bindAddress: 0.0.0.0
database:
  type: postgres
  host: localhost
  port: 5432
  user: gobe
  password: secure
```

---

## **Roadmap**

- [x] Modularização total e interfaces plugáveis
- [x] Zero-config com geração automática de certificados
- [x] Integração com keyring do sistema
- [x] API REST para autenticação e gerenciamento
- [x] Autenticação via certificados e senhas seguras
- [x] CLI para gerenciamento e monitoramento
- [x] Integração com `gdbase` para gerenciamento de bancos via Docker
- [–] Suporte a múltiplos bancos de dados (Parcial concluído)
- [&nbsp;&nbsp;] Integração com Prometheus para monitoramento
- [&nbsp;&nbsp;] Suporte a middlewares personalizados
- [&nbsp;&nbsp;] Integração com Grafana para visualização de métricas
- [–] Documentação completa e exemplos de uso (Parcial concluído)
- [–] Testes automatizados e CI/CD (Parcial concluído)

---

## **Contributing**

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou enviar pull requests. Veja o [Guia de Contribuição](docs/CONTRIBUTING.md) para mais detalhes.

---

## **Contact**

💌 **Developer**:  
[Rafael Mori](mailto:rafa-mori@gmail.com)  
💼 [Follow me on GitHub](https://github.com/rafa-mori)  
Estou aberto a colaborações e novas ideias. Se achou o projeto interessante, entre em contato!


