# IonBus

**IonBus** is a high-performance, lightweight IoT telemetry system designed to move data from distributed devices to backend services with speed and reliability.  
Built with **Go**, **MQTT**, **RabbitMQ**, and **PostgreSQL**, IonBus provides a flexible foundation for real-time sensor data ingestion, processing, and persistence.

## Overview

IonBus bridges the gap between **IoT devices** and **data infrastructure**.  
It enables devices — such as Raspberry Pis, embedded sensors, or industrial controllers — to securely publish telemetry data over MQTT, while backend services such as RabbitMQ handle aggregation, alerting, and long-term storage.

This system emphasizes:
- **High throughput** with minimal overhead  
- **Reliable message delivery** and retry handling  
- **Extensible architecture** for analytics or monitoring
- **Developer-friendly design** for rapid iteration and testing  

### Architecture Components

| Component | Description |
|------------|-------------|
| **Device Layer** | IoT or edge devices (e.g., Raspberry Pis, sensors, or embedded clients) publish telemetry data such as temperature, humidity, or motion readings via **MQTT**. |
| **Local MQTT Broker** | A lightweight broker (e.g., **Mosquitto** or **EMQX**) running on-prem or edge devices that receives and routes telemetry messages with minimal latency. |
| **Bridge Service (MQTT → RabbitMQ)** | A Go service that subscribes to MQTT topics and republishes messages to **RabbitMQ**. Provides decoupling, buffering, and reliable delivery to backend services. |
| **RabbitMQ Broker** | Acts as the **intermediary layer** between local ingestion and backend processing. Handles queueing, fanout, dead-lettering, and ensures messages are not lost. |
| **Processing Layer** | Backend workers (Go services) consume messages from RabbitMQ, trigger alerts or aggregate data for analytics. |
| **Storage Layer** | **PostgreSQL** or **TimescaleDB** stores processed telemetry for long-term query, reporting, and visualization. |
| **API / Visualization Layer** *(optional)* | Exposes REST or WebSocket endpoints for dashboards, monitoring, or analytics tools (e.g., Grafana). |
| **Infrastructure** | Managed with **Docker Compose**, orchestrating MQTT, RabbitMQ, backend services, and database in a reproducible environment. |

