# Minimal Docker Orchestrator in Go

This project is a lightweight container orchestrator written in Go, inspired by basic Kubernetes concepts. It allows you to:

- Define pods with Docker images and resource requirements
- Automatically schedule pods to available nodes
- Start corresponding containers via the Docker API
- Clean up running containers gracefully on shutdown
- Expose a simple REST API to manage pods

---

## Features

-  **Automatic pod scheduling** to nodes based on available CPU and memory
-  **Node agent** that starts Docker containers when pods are assigned
-  **Graceful container cleanup** on `SIGINT` (Ctrl+C)
-  **HTTP REST API**:
  - `POST /pods`: Create a new pod
  - `GET /pods`: List all created pods
  - `DELETE /pods/:name`: Delete the container with the indicated name

---

## Getting Started

### Requirements
Check `go.mod` to see the requirements and run `go mod tidy`
- [Go](https://golang.org/dl/) 1.23 used
- [Docker](https://www.docker.com/) 28.3.2 used

### Clone and Run

```bash
git clone https://github.com/AndrePoix/mini-kubernetes.git
cd mini-kubernetes
go run .
