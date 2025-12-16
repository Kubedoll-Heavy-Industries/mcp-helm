variable "IMAGE" {
  default = "mcp-helm"
}

variable "TAG" {
  default = "dev"
}

group "default" {
  targets = ["local"]
}

target "_common" {
  context = "."
  dockerfile = "Dockerfile"
  args = {
    VERSION = "dev"
    COMMIT  = "none"
    DATE    = "unknown"
  }
  labels = {
    "org.opencontainers.image.source" = "https://github.com/Kubedoll-Heavy-Industries/mcp-helm"
  }
}

target "local" {
  inherits = ["_common"]
  target = "runtime"
  tags = ["${IMAGE}:${TAG}"]
  output = ["type=docker"]
  cache-from = ["type=gha"]
  cache-to = ["type=gha,mode=max"]
}

target "debug" {
  inherits = ["_common"]
  target = "debug"
  tags = ["${IMAGE}:${TAG}-debug"]
  output = ["type=docker"]
  cache-from = ["type=gha"]
  cache-to = ["type=gha,mode=max"]
}

target "release" {
  inherits = ["_common"]
  target = "runtime"
  platforms = ["linux/amd64", "linux/arm64"]
  tags = ["${IMAGE}:${TAG}"]
  cache-from = ["type=gha"]
  cache-to = ["type=gha,mode=max"]
}
