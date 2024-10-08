## Monitoring Stack: Prometheus, Grafana, and Alertmanager
This repository provides a Docker Compose setup for a monitoring stack including Prometheus, Grafana, and Alertmanager. This stack allows you to monitor and visualize metrics from your applications and infrastructure, and receive alerts based on defined rules.

# Prerequisites
Docker
Docker Compose

# Getting Started

1. Clone the Repository

 - git clone https://github.com/jaycynth/monitoring-performance-metrics-in-containerized-infrastructure.git
- cd monitoring-stack

2. Configuration

    - Environment Variables
    - ALERT_EMAIL_TO=your_email@example.com
    - ALERT_EMAIL_FROM=alertmanager@example.com
    - ALERT_SMARTHOST=smtp.example.com:587
    - ALERT_SMTP_USERNAME=your_smtp_user
    - ALERT_SMTP_PASSWORD=your_smtp_password


- Prometheus Configuration
Modify prometheus/prometheus.yml to configure Prometheus scrape targets and alerting rules.

- Alertmanager Configuration
Modify alertmanager/config.yml to configure Alertmanager routing and receivers. It uses environment variables for sensitive data.

3. Build and Start the Stack
Use the provided Makefile to build and start the containers:

- make build    # Build Docker images
- make up       # Start the containers

4. Access the Services
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (default login: admin / admin)
- Alertmanager: http://localhost:9093

5. Configure Grafana
- Add Data Source: Configure Prometheus as a data source.
- Import Dashboards: You can use predefined dashboards or create custom ones.


6. Viewing Logs
To view logs from the containers, use:

- make logs
- make logs-prometheus

7. Stopping and Cleaning Up
To stop the containers:
- make down
To clean up unused Docker resources:
- make clean

# Alerting
Configure alerting rules in prometheus/alert.rules.yml and set up routing in alertmanager/config.yml. Alerts can be sent via email or other notification systems supported by Alertmanager.
