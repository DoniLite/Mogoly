package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
)

// DatabaseType représente les types de bases de données supportées
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	MongoDB    DatabaseType = "mongodb"
	Redis      DatabaseType = "redis"
	MariaDB    DatabaseType = "mariadb"
)

// DatabaseInstance représente une instance de base de données
type DatabaseInstance struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         DatabaseType `json:"type"`
	ContainerID  string       `json:"container_id"`
	Port         int          `json:"port"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	DatabaseName string       `json:"database_name"`
	Status       string       `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	ExternalPort int          `json:"external_port"`
}

// DatabaseConfig contient la configuration pour créer une DB
type DatabaseConfig struct {
	Type         DatabaseType `json:"type"`
	Name         string       `json:"name"`
	Username     string       `json:"username,omitempty"`
	Password     string       `json:"password,omitempty"`
	DatabaseName string       `json:"database_name,omitempty"`
	Version      string       `json:"version,omitempty"`
}

// CloudDBManager gère les instances de bases de données
type CloudDBManager struct {
	dockerClient *client.Client
	instances    map[string]*DatabaseInstance
	portCounter  int
	networkName  string
}

// NewCloudDBManager crée un nouveau gestionnaire
func NewCloudDBManager() (*CloudDBManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("erreur création client Docker: %v", err)
	}

	manager := &CloudDBManager{
		dockerClient: cli,
		instances:    make(map[string]*DatabaseInstance),
		portCounter:  5432, // Port de base
		networkName:  "clouddb-network",
	}

	// Créer le réseau Docker si nécessaire
	if err := manager.ensureNetwork(); err != nil {
		return nil, fmt.Errorf("erreur création réseau: %v", err)
	}

	return manager, nil
}

// ensureNetwork crée le réseau Docker si il n'existe pas
func (m *CloudDBManager) ensureNetwork() error {
	ctx := context.Background()
	
	networks, err := m.dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	// Vérifier si le réseau existe déjà
	for _, net := range networks {
		if net.Name == m.networkName {
			return nil
		}
	}

	// Créer le réseau
	_, err = m.dockerClient.NetworkCreate(ctx, m.networkName, network.CreateOptions{
		Driver: "bridge",
	})
	return err
}

// generateRandomString génère une chaîne aléatoire
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// getNextPort retourne le prochain port disponible
func (m *CloudDBManager) getNextPort() int {
	m.portCounter++
	return m.portCounter
}

// getDockerConfig retourne la configuration Docker pour un type de DB
func (m *CloudDBManager) getDockerConfig(config DatabaseConfig) (string, map[string]string, []string, error) {
	var image string
	var env map[string]string
	var exposedPorts []string

	// Générer un mot de passe aléatoire si non fourni
	password := config.Password
	if password == "" {
		password = generateRandomString(16)
	}

	switch config.Type {
	case PostgreSQL:
		version := config.Version
		if version == "" {
			version = "15"
		}
		image = fmt.Sprintf("postgres:%s", version)
		env = map[string]string{
			"POSTGRES_DB":       config.DatabaseName,
			"POSTGRES_USER":     config.Username,
			"POSTGRES_PASSWORD": password,
		}
		exposedPorts = []string{"5432/tcp"}

	case MySQL:
		version := config.Version
		if version == "" {
			version = "8.0"
		}
		image = fmt.Sprintf("mysql:%s", version)
		env = map[string]string{
			"MYSQL_DATABASE":      config.DatabaseName,
			"MYSQL_USER":          config.Username,
			"MYSQL_PASSWORD":      password,
			"MYSQL_ROOT_PASSWORD": generateRandomString(16),
		}
		exposedPorts = []string{"3306/tcp"}

	case MariaDB:
		version := config.Version
		if version == "" {
			version = "10.9"
		}
		image = fmt.Sprintf("mariadb:%s", version)
		env = map[string]string{
			"MARIADB_DATABASE":      config.DatabaseName,
			"MARIADB_USER":          config.Username,
			"MARIADB_PASSWORD":      password,
			"MARIADB_ROOT_PASSWORD": generateRandomString(16),
		}
		exposedPorts = []string{"3306/tcp"}

	case MongoDB:
		version := config.Version
		if version == "" {
			version = "6.0"
		}
		image = fmt.Sprintf("mongo:%s", version)
		env = map[string]string{
			"MONGO_INITDB_DATABASE":      config.DatabaseName,
			"MONGO_INITDB_ROOT_USERNAME": config.Username,
			"MONGO_INITDB_ROOT_PASSWORD": password,
		}
		exposedPorts = []string{"27017/tcp"}

	case Redis:
		version := config.Version
		if version == "" {
			version = "7"
		}
		image = fmt.Sprintf("redis:%s", version)
		env = map[string]string{}
		exposedPorts = []string{"6379/tcp"}
		
		// Redis avec authentification
		if password != "" {
			env["REDIS_PASSWORD"] = password
		}

	default:
		return "", nil, nil, fmt.Errorf("type de base de données non supporté: %s", config.Type)
	}

	return image, env, exposedPorts, nil
}

// CreateInstance crée une nouvelle instance de base de données
func (m *CloudDBManager) CreateInstance(config DatabaseConfig) (*DatabaseInstance, error) {
	ctx := context.Background()

	// Valeurs par défaut
	if config.Username == "" {
		config.Username = "admin"
	}
	if config.DatabaseName == "" {
		config.DatabaseName = config.Name
	}

	// Obtenir la configuration Docker
	image, envVars, exposedPorts, err := m.getDockerConfig(config)
	if err != nil {
		return nil, err
	}

	// Créer l'instance
	instanceID := generateRandomString(12)
	instance := &DatabaseInstance{
		ID:           instanceID,
		Name:         config.Name,
		Type:         config.Type,
		Username:     config.Username,
		Password:     config.Password,
		DatabaseName: config.DatabaseName,
		Status:       "creating",
		CreatedAt:    time.Now(),
		ExternalPort: m.getNextPort(),
	}

	// Configuration du conteneur
	containerConfig := &container.Config{
		Image: image,
		Env:   make([]string, 0, len(envVars)),
	}

	// Ajouter les variables d'environnement
	for key, value := range envVars {
		containerConfig.Env = append(containerConfig.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Configuration des ports
	portBindings := nat.PortMap{}
	exposedPortsMap := nat.PortSet{}
	
	for _, port := range exposedPorts {
		natPort, _ := nat.NewPort("tcp", strings.Split(port, "/")[0])
		exposedPortsMap[natPort] = struct{}{}
		portBindings[natPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: strconv.Itoa(instance.ExternalPort),
			},
		}
	}

	containerConfig.ExposedPorts = exposedPortsMap

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			m.networkName: {},
		},
	}

	// Créer le conteneur
	containerName := fmt.Sprintf("clouddb-%s-%s", config.Type, instanceID)
	resp, err := m.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("erreur création conteneur: %v", err)
	}

	instance.ContainerID = resp.ID

	// Démarrer le conteneur
	if err := m.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("erreur démarrage conteneur: %v", err)
	}

	instance.Status = "running"
	m.instances[instanceID] = instance

	return instance, nil
}

// GetInstance récupère une instance par son ID
func (m *CloudDBManager) GetInstance(id string) (*DatabaseInstance, bool) {
	instance, exists := m.instances[id]
	return instance, exists
}

// ListInstances retourne toutes les instances
func (m *CloudDBManager) ListInstances() []*DatabaseInstance {
	instances := make([]*DatabaseInstance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}
	return instances
}

// DeleteInstance supprime une instance
func (m *CloudDBManager) DeleteInstance(id string) error {
	instance, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance non trouvée: %s", id)
	}

	ctx := context.Background()

	// Arrêter et supprimer le conteneur
	if err := m.dockerClient.ContainerStop(ctx, instance.ContainerID, container.StopOptions{}); err != nil {
		log.Printf("Erreur arrêt conteneur: %v", err)
	}

	if err := m.dockerClient.ContainerRemove(ctx, instance.ContainerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("erreur suppression conteneur: %v", err)
	}

	delete(m.instances, id)
	return nil
}

// API Handlers

func (m *CloudDBManager) createInstanceHandler(w http.ResponseWriter, r *http.Request) {
	var config DatabaseConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	instance, err := m.CreateInstance(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

func (m *CloudDBManager) getInstanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	instance, exists := m.GetInstance(id)
	if !exists {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

func (m *CloudDBManager) listInstancesHandler(w http.ResponseWriter, r *http.Request) {
	instances := m.ListInstances()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

func (m *CloudDBManager) deleteInstanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := m.DeleteInstance(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *CloudDBManager) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	_, err := NewCloudDBManager()
	if err != nil {
		log.Fatal("Erreur initialisation:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Serveur démarré sur le port %s\n", port)
	fmt.Println("API disponible sur http://localhost:" + port + "/api/v1/")
	
}