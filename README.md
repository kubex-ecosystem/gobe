# GoBE - Modular & Secure Back-end

![GoBE Banner](docs/assets/top_banner_lg_b.png)

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/rafa-mori/gobe/blob/main/LICENSE)
[![Automation](https://img.shields.io/badge/automation-zero%20config-blue)](#features)
[![Modular](https://img.shields.io/badge/modular-yes-yellow)](#features)
[![Security](https://img.shields.io/badge/security-high-red)](#features)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://github.com/rafa-mori/gobe/blob/main/CONTRIBUTING.md)
[![Build](https://github.com/rafa-mori/gobe/actions/workflows/kubex_go_release.yml/badge.svg)](https://github.com/rafa-mori/gobe/actions/workflows/kubex_go_release.yml)

---

**A modular, secure, and zero-config backend for modern Go applications.**

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

GoBE is a modular backend developed in Go, focused on **security, automation, and flexibility**. It can run as a **main server** or be used **as a module** for managing features like **encryption, certificates, middlewares, logging, and authentication**.

### **Current Status**

- **Zero-config:** No manual configuration required, generates all certificates and securely stores sensitive information in the system keyring.
- **Extensible:** Can be integrated with other systems or run standalone.
- **Modularization:** The project is fully modular, with all logic encapsulated in well-defined interfaces.
- **Integration with `gdbase`:** Database management is handled via Docker, allowing for easy setup and optimization.
- **REST API:** Provides endpoints for authentication, user management, products, clients, and cronjobs.
- **Authentication:** Uses dynamically generated certificates, random passwords, and secure keyring for robust security.
- **CLI:** A powerful command-line interface for managing the server, including commands to start, stop, and monitor services.
- **Logging and Security Management:** Protected routes, secure storage, and request monitoring are implemented to ensure data integrity and security.
- **Multi-database support:** Currently supports PostgreSQL and SQLite, with plans for more databases in the future.
- **Prometheus and Grafana integration:** Planned for monitoring and metrics visualization.
- **Documentation:** Continuous improvement to provide comprehensive documentation for all endpoints and functionalities.
- **Unit Tests:** While all functionalities are operational, unit tests are being developed to ensure reliability and robustness.
- **CI/CD:** Automated tests and continuous integration are in progress to maintain code quality and deployment efficiency.
- **Complete Documentation:** The documentation is being expanded to cover all aspects of the project, including usage examples and detailed endpoint descriptions.
- **Automated Tests:** Although the functionalities are implemented, unit tests are being developed to ensure reliability and robustness.

## **Project Evolution**

The project has undergone significant evolution since its inception. Initially focused on basic functionalities, it has now expanded to include a wide range of features that enhance security, modularity, and ease of use.
The current version of GoBE is a result of continuous improvements and refinements, with a strong emphasis on security and automation. The system is designed to be user-friendly, allowing developers to focus on building applications without worrying about backend complexities.
The modular architecture allows for easy integration with other systems, making GoBE a versatile choice for modern Go applications. The project is actively maintained, with ongoing efforts to enhance its capabilities and ensure it meets the evolving needs of developers.

Documentation and CI/CD are key focus areas for the next updates

---

## **Features**

✨ **Fully modular**

- All logic follows well-defined interfaces, ensuring encapsulation.
- Can be used as a server or as a library/module.

🔒 **Zero-config, but customizable**

- Runs without initial configuration, but supports customization via files.
- Automatically generates certificates, passwords, and secure settings.

🔗 **Direct integration with `gdbase`**

- Database management via Docker.
- Automatic optimizations for persistence and performance.

🛡️ **Advanced authentication**

- Dynamically generated certificates.
- Random passwords and secure keyring.

🌐 **Robust REST API**

- Endpoints for authentication, user management, products, clients, and cronjobs.

📋 **Log and security management**

- Protected routes, secure storage, and request monitoring.

🧑‍💻 **Powerful CLI**

- Commands to start, configure, and monitor the server.

---

## **Installation**

Requirements:

- Go 1.19+
- Docker (for database integration via gdbase)

Clone the repository and build GoBE:

```sh
# Clone the repository
git clone https://github.com/rafa-mori/gobe.git
cd gobe
go build -o gobe .
```

---

## **Usage**

### CLI

Start the main server:

```sh
./gobe start -p 3666 -b "0.0.0.0"
```

This starts the server, generates certificates, sets up databases, and begins listening for requests!

See all available commands:

```sh
./gobe --help
```

**Main commands:**

| Command   | Function                                         |
|-----------|--------------------------------------------------|
| `start`   | Starts the server                                |
| `stop`    | Safely stops the server                          |
| `restart` | Restarts all services                            |
| `status`  | Shows the status of the server and active services|
| `config`  | Generates an initial configuration file          |
| `logs`    | Displays server logs                             |

---

### Configuration

GoBE can run without any initial configuration, but supports customization via YAML/JSON files. By default, everything is generated automatically on first use.

Example configuration:

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

#### Messaging Integrations

WhatsApp and Telegram bots can be configured via the `config/discord_config.json` file under the `integrations` section:

```json
{
  "integrations": {
    "whatsapp": {
      "enabled": true,
      "access_token": "<token>",
      "verify_token": "<verify>",
      "phone_number_id": "<number>",
      "webhook_url": "https://your.server/whatsapp/webhook"
    },
    "telegram": {
      "enabled": true,
      "bot_token": "<bot token>",
      "webhook_url": "https://your.server/telegram/webhook",
      "allowed_updates": ["message", "callback_query"]
    }
  }
}
```

After setting up the file or environment variables, the server will expose the following endpoints:

- `POST /api/v1/whatsapp/send` and `/api/v1/whatsapp/webhook`
- `POST /api/v1/telegram/send` and `/api/v1/telegram/webhook`

Each route also provides a `/ping` endpoint for health checks.

---

## **Roadmap**

- [x] Full modularization and pluggable interfaces
- [x] Zero-config with automatic certificate generation
- [x] Integration with system keyring
- [x] REST API for authentication and management
- [x] Authentication via certificates and secure passwords
- [x] CLI for management and monitoring
- [x] Integration with `gdbase` for database management via Docker
- [–] Multi-database support (Partially completed)
- [  ] Prometheus integration for monitoring
- [  ] Support for custom middlewares
- [  ] Grafana integration for metrics visualization
- [–] Complete documentation and usage examples (Partially completed)
- [–] Automated tests and CI/CD (Partially completed)

---

## **Contributing**

Contributions are welcome! Feel free to open issues or submit pull requests. See the [Contribution Guide](docs/CONTRIBUTING.md) for more details.

---

## **Contact**

💌 **Developer**:  
[Rafael Mori](mailto:faelmori@gmail.com)  
💼 [Follow me on GitHub](https://github.com/rafa-mori)  
I'm open to collaborations and new ideas. If you found the project interesting, get in touch!


