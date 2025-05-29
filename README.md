# GDBase - Back-end Modular & Seguro

# 🚀 **GoBE - Back-end Modular & Seguro**  

## 🔥 **Visão Geral**  
**GoBE** é um back-end modular desenvolvido em Go, focado em **segurança, automação e flexibilidade**. Ele pode rodar como **servidor principal** ou ser utilizado **como módulo** para gerenciamento de funcionalidades específicas como **criptografia, certificados, middlewares, logging e autenticação**.  

Ele **não exige configuração manual**, gerando todos os certificados e armazenando informações sensíveis de forma segura no **keyring do sistema**.  

## 🔗 **Recursos Principais**  
✅ **Totalmente modular** → Todas as lógicas seguem interfaces bem definidas, garantindo encapsulamento.  
✅ **Zero-config, mas personalizável** → Pode rodar sem configuração inicial ou ser ajustado via arquivos.  
✅ **Integração direta com `gdbase`** → Gerenciamento de bancos de dados via Docker e otimizações automáticas.  
✅ **Autenticação avançada** → **Certificados gerados dinamicamente, senhas aleatórias e keyring seguro**.  
✅ **API REST robusta** → Endpoints para **autenticação, gerenciamento de usuários, produtos, clientes e cronjobs**.  
✅ **Gerenciamento de logs e segurança** → **Rotas protegidas**, armazenamento seguro e monitoramento de requisições.  
✅ **CLI poderosa** → Com comandos para iniciar, configurar e monitorar o servidor.  

## 📝 **Instalação**  
Clone o repositório e compile o GoBE:  

```sh
git clone https://github.com/rafa-mori/gobe.git
cd gobe
go build -o gobe .
```

## 🚀 **Rodando o Servidor**  
Para iniciar o **GoBE**, basta rodar:  

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

Isso **inicializa o servidor, gera certificados**, configura bancos de dados e começa a escutar requisições!  

## 🔎 **Comandos da CLI**  
O GoBE possui comandos internos para facilitar o gerenciamento:  

```sh
./gobe --help
```

💡 **Aqui está um resumo dos comandos disponíveis:**  

| Comando            | Função                                             |
|--------------------|----------------------------------------------------|
| `start`           | Inicializa o servidor                              |
| `stop`            | Encerra o servidor de forma segura                 |
| `restart`         | Reinicia todos os serviços                         |
| `status`          | Exibe o status do servidor e dos serviços ativos   |
| `config`          | Gera um arquivo de configuração inicial            |
| `logs`            | Exibe os logs do servidor                          |


