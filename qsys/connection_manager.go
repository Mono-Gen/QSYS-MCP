package qsys

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"qsys-mcp/config"
)

type CoreEntry struct {
	Config CoreConfig
	Qrc    *QrcClient
	Ecp    *EcpClient
}

type ConnectionManager struct {
	entries      map[CoreAlias]*CoreEntry
	defaultAlias CoreAlias
	mutex        sync.RWMutex
}

var CurrentConnectionManager *ConnectionManager

func InitConnectionManager() error {
	envVar := os.Getenv("QSYS_CORES")
	cm := &ConnectionManager{
		entries: make(map[CoreAlias]*CoreEntry),
	}
	
	if err := cm.parseConfig(envVar); err != nil {
		return err
	}
	
	CurrentConnectionManager = cm
	return nil
}

func (cm *ConnectionManager) parseConfig(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	segments := strings.Split(raw, ",")
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		eqIdx := strings.Index(segment, "=")
		if eqIdx == -1 {
			return fmt.Errorf("invalid QSYS_CORES entry %q: expected format alias=host[:qrcPort[:ecpPort]]", segment)
		}

		alias := CoreAlias(strings.TrimSpace(segment[:eqIdx]))
		rest := strings.TrimSpace(segment[eqIdx+1:])

		if alias == "" {
			return fmt.Errorf("invalid QSYS_CORES entry %q: alias cannot be empty", segment)
		}
		if rest == "" {
			return fmt.Errorf("invalid QSYS_CORES entry %q: host cannot be empty", segment)
		}

		parts := strings.Split(rest, ":")
		host := parts[0]
		qrcPort := DefaultQrcPort
		ecpPort := 1702
		var user, pin string

		if len(parts) > 1 && parts[1] != "" {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid qrcPort in QSYS_CORES entry %q: %w", segment, err)
			}
			qrcPort = port
		}

		if len(parts) > 2 && parts[2] != "" {
			port, err := strconv.Atoi(parts[2])
			if err != nil {
				return fmt.Errorf("invalid ecpPort in QSYS_CORES entry %q: %w", segment, err)
			}
			ecpPort = port
		}

		if len(parts) > 3 {
			user = parts[3]
		}
		if len(parts) > 4 {
			pin = parts[4]
		}

		if host == "" {
			return fmt.Errorf("invalid QSYS_CORES entry %q: host cannot be empty", segment)
		}

		qrcClient := NewQrcClient(host, qrcPort)
		if user != "" || pin != "" {
			qrcClient.SetAuth(user, pin)
		}
		ecpClient := NewEcpClient(host, ecpPort)

		cm.entries[alias] = &CoreEntry{
			Config: CoreConfig{
				Alias:   alias,
				Host:    host,
				QrcPort: qrcPort,
				EcpPort: ecpPort,
				User:    user,
				Pin:     pin,
			},
			Qrc: qrcClient,
			Ecp: ecpClient,
		}
	}

	if len(cm.entries) == 1 {
		for alias := range cm.entries {
			cm.defaultAlias = alias
			break
		}
	}

	return nil
}

func (cm *ConnectionManager) resolveAlias(alias CoreAlias) (CoreAlias, error) {
	if alias != "" {
		return alias, nil
	}

	if cm.defaultAlias != "" {
		return cm.defaultAlias, nil
	}

	if len(cm.entries) == 0 {
		return "", errors.New("no Cores configured. Set QSYS_CORES environment variable")
	}

	var keys []string
	for k := range cm.entries {
		keys = append(keys, string(k))
	}
	return "", fmt.Errorf("multiple Cores configured — alias is required. Use one of: %s", strings.Join(keys, ", "))
}

func (cm *ConnectionManager) GetClients(alias CoreAlias) (*QrcClient, *EcpClient, error) {
	cm.mutex.RLock()
	resolved, err := cm.resolveAlias(alias)
	if err != nil {
		cm.mutex.RUnlock()
		return nil, nil, err
	}

	entry, ok := cm.entries[resolved]
	cm.mutex.RUnlock()

	if !ok {
		var keys []string
		cm.mutex.RLock()
		for k := range cm.entries {
			keys = append(keys, string(k))
		}
		cm.mutex.RUnlock()
		return nil, nil, fmt.Errorf("unknown Core alias %q. Configured cores: %s", resolved, strings.Join(keys, ", "))
	}

	// Lazy connect
	if !entry.Qrc.IsConnected() {
		config.DebugLog("connecting QRC for %q", resolved)
		if err := entry.Qrc.Connect(); err != nil {
			logError("QRC connect failed for %q: %v", resolved, err)
		}
	}

	if !entry.Ecp.IsConnected() {
		config.DebugLog("connecting ECP for %q", resolved)
		if err := entry.Ecp.Connect(); err != nil {
			logError("ECP connect failed for %q: %v", resolved, err)
		}
	}

	return entry.Qrc, entry.Ecp, nil
}

func (cm *ConnectionManager) ListCores() []map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var list []map[string]interface{}
	for alias, entry := range cm.entries {
		state := Disconnected
		if entry.Qrc.IsConnected() || entry.Ecp.IsConnected() {
			state = Connected
		}

		list = append(list, map[string]interface{}{
			"alias":     string(alias),
			"host":      entry.Config.Host,
			"state":     string(state),
			"isDefault": alias == cm.defaultAlias,
		})
	}
	return list
}

func (cm *ConnectionManager) DisconnectAll() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, entry := range cm.entries {
		_ = entry.Qrc.Disconnect()
		_ = entry.Ecp.Disconnect()
	}
}

func logError(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "[qsys] "+format+"\n", v...)
}
