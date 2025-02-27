package docker

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// DockerConfig 定义 Docker 守护进程配置结构
type DockerConfig struct {
    Proxies         map[string]string `json:"proxies"`          // HTTP/HTTPS 代理
    Mirrors         []string          `json:"mirrors"`          // 镜像加速器
    RegistryMirrors []string          `json:"registry-mirrors"` // Docker 原生镜像加速器配置
}

func ReadDaemonConfig() (*DockerConfig, error) {
    configPath := "/etc/docker/daemon.json"
    config := &DockerConfig{
        Proxies: make(map[string]string),
        Mirrors: make([]string, 0),
    }

    // 如果配置文件不存在，返回空配置
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return config, nil
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }

    var daemonConfig struct {
        RegistryMirrors []string `json:"registry-mirrors,omitempty"`
        HTTPProxy       string   `json:"http-proxy,omitempty"`
        HTTPSProxy      string   `json:"https-proxy,omitempty"`
        NoProxy         string   `json:"no-proxy,omitempty"`
    }

    if err := json.Unmarshal(data, &daemonConfig); err != nil {
        return nil, err
    }

    // 转换为我们的配置格式
    config.RegistryMirrors = daemonConfig.RegistryMirrors
    config.Proxies["http"] = daemonConfig.HTTPProxy
    config.Proxies["https"] = daemonConfig.HTTPSProxy
    config.Proxies["no"] = daemonConfig.NoProxy
    config.Mirrors = daemonConfig.RegistryMirrors

    return config, nil
}

func UpdateDaemonConfig(config *DockerConfig) error {
    // 转换为 daemon.json 格式
    daemonConfig := struct {
        RegistryMirrors []string `json:"registry-mirrors,omitempty"`
        HTTPProxy       string   `json:"http-proxy,omitempty"`
        HTTPSProxy      string   `json:"https-proxy,omitempty"`
        NoProxy         string   `json:"no-proxy,omitempty"`
    }{
        RegistryMirrors: config.Mirrors,
        HTTPProxy:       config.Proxies["http"],
        HTTPSProxy:      config.Proxies["https"],
        NoProxy:         config.Proxies["no"],
    }

    data, err := json.MarshalIndent(daemonConfig, "", "    ")
    if err != nil {
        return err
    }

    configPath := "/etc/docker/daemon.json"
    if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
        return err
    }

    return os.WriteFile(configPath, data, 0644)
}