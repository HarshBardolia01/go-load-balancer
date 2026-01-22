package server

import (
	"fmt"
)

func (s *Server) RunCLI() {
	for {
		stop := s.PrintMenu()
		if stop {
			break
		}
	}
}

func (s *Server) PrintMenu() bool {
	heartbeatStatus := "OFF"
	if s.GetIsHeartbeatRunning() {
		heartbeatStatus = "ON"
	}

	fmt.Printf("\n======================================\n")
	fmt.Printf("Server: %s, Heartbeat: %s\n", s.InstanceId, heartbeatStatus)
	fmt.Printf("\n======================================\n")
	fmt.Println("1. Get Server Info")
	fmt.Println("2. Get Server Stats")
	fmt.Println("3. Update Server Weight")
	fmt.Println("4. Update Server Type")
	fmt.Println("5. Start Heartbeat")
	fmt.Println("6. Stop Heartbeat")
	fmt.Println("7. Exit")
	fmt.Print("Enter your choice: ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		s.GetServerInfo()
	case 2:
		s.GetServerStats()
	case 3:
		fmt.Printf("Enter new weight (greater than or equal to 0): ")
		var weight int
		fmt.Scanln(&weight)

		if weight < 0 {
			fmt.Println("Invalid weight. Weight must be greater than or equal to 0")
			break
		}

		s.SetWeight(weight)
	case 4:
		fmt.Printf("Enter new server type (fast/slow/medium): ")
		var serverType string
		fmt.Scanln(&serverType)

		if serverType != serverTypeFast && serverType != serverTypeSlow && serverType != serverTypeMedium {
			fmt.Println("Invalid server type. Server type must be one of fast/slow/medium")
			break
		}

		s.SetServerType(serverType)
	case 5:
		s.StartHeartbeat()
	case 6:
		s.StopHeartbeat()
	case 7:
		fmt.Printf("Are you sure you want to exit? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "y" || confirm == "Y" {
			return true
		}
	default:
		fmt.Println("Invalid choice")
	}

	return false
}

func (s *Server) GetServerInfo() {
	fmt.Printf("\n======================================\n")
	fmt.Printf("Server Info\n")
	fmt.Printf("======================================\n")
	fmt.Printf("Schema: %s\n", s.Schema)
	fmt.Printf("Host: %s\n", s.Host)
	fmt.Printf("Port: %d\n", s.Port)
	fmt.Printf("Instance ID: %s\n", s.InstanceId)
	fmt.Printf("Weight: %d\n", s.Weight)
	fmt.Printf("Server Type: %s\n", s.ServerType)
	fmt.Printf("======================================\n")
}

func (s *Server) GetServerStats() {
	fmt.Printf("\n======================================\n")
	fmt.Printf("Server Stats\n")
	fmt.Printf("======================================\n")
	fmt.Printf("Active Connections: %d\n", s.GetActiveConnections())
	fmt.Printf("Total Requests Served: %d\n", s.GetTotalRequestsServed())
	fmt.Printf("======================================\n")
}
