job "arangodb-agent-economique" {
  datacenters = ["dc1"]
  type = "service"

  group "arangodb" {
    count = 1

    network {
      port "http" {
        static = 8529
        to     = 8529
      }
    }

    task "arangodb" {
      driver = "docker"

      config {
        image = "arangodb:3.11"
        ports = ["http"]
        
        volumes = [
          "/opt/arangodb_data:/var/lib/arangodb3"
        ]
        
        # Arguments pour configurer ArangoDB
        args = [
          "--server.endpoint", "tcp://0.0.0.0:8529",
          "--server.authentication", "true",
          "--log.level", "info"
        ]
      }

      env {
        ARANGO_ROOT_PASSWORD = "agent_economique_2025"
      }

      resources {
        cpu    = 500
        memory = 1024
      }

      # Service registry désactivé (pas de Consul)
      # ArangoDB accessible directement sur 10.0.0.1:8529
    }
  }
}
