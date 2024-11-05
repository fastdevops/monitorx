package api

import (
	"fmt"
	"github.com/fastdevops/monitorx/global"
	"github.com/fastdevops/monitorx/logger"
	"github.com/fastdevops/monitorx/monitor"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func DashboardHandler(c *gin.Context) {
	// 获取监控数据
	hadoop, dorisFE, dorisBE, kafka, dolphinscheduler, zookeeper, resourceStatus := CheckStartApi()

	// 模板数据传递
	data := gin.H{
		"title":            "大数据监控面板",
		"hadoopStatus":     hadoop,
		"dorisFEStatus":    dorisFE,
		"dorisBEStatus":    dorisBE,
		"kafkaStatus":      kafka,
		"dolphinscheduler": dolphinscheduler,
		"zookeeperStatus":  zookeeper,
		"resourceStatus":   resourceStatus,
	}

	c.HTML(http.StatusOK, "dashboard.html", data)
}

// Helper functions to filter the submodules by category (Hadoop, Flink, Yarn)
func resourceStatusFilter(resources []ComponentStatus, module string) []ComponentStatus {
	var filtered []ComponentStatus
	for _, resource := range resources {
		if strings.Contains(resource.Name, module) {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

type ComponentStatus struct {
	Name string `json:"name"`
	RAM  string `json:"ram"`
	ROM  string `json:"rom"`
	CPU  string `json:"cpu"`
}

func CheckErr(err error) {
	if err != nil {
		logger.Logger.Error("Error occurred: ", zap.Error(err))
	}
}

func CheckStartApi() (hadoop, dorisFE, dorisBE, kafka, dolphinscheduler, zookeeper bool, resourceStatus []ComponentStatus) {
	// Initialize configuration
	cf := global.Config

	// Initialize all service statuses to true (running)
	hadoop, dorisFE, dorisBE, kafka, dolphinscheduler, zookeeper = true, true, true, true, true, true

	// Initialize resource status list
	var HostResourceStatus []ComponentStatus

	// Check all nodes
	for _, node := range cf.Nodes {
		HostserviceStatus, err := monitor.CheckServices(node)
		CheckErr(err)

		cpuUsage, memUsage, diskUsage, err := monitor.CheckNodeResources(node)
		CheckErr(err)

		logger.Logger.Info(fmt.Sprintf("Node: %s, CPU: %s, RAM: %s, Disk: %s", node.Name, cpuUsage, memUsage, diskUsage))

		// Distinguish services and update their status
		ErrorServices := []string{}
		for _, ConfigService := range node.Services {
			for _, check_service := range HostserviceStatus {
				if ConfigService == check_service {
					continue
				} else {
					ErrorServices = append(ErrorServices, ConfigService)
				}
			}

			if ErrorServices != nil {
				switch ConfigService {
				case "NameNode", "DFSZKFailoverController", "ResourceManager", "JournalNode", "DataNode", "NodeManager":
					hadoop = false
				case "DorisFE":
					dorisFE = false
				case "DorisBE":
					dorisBE = false
				case "Kafka":
					kafka = false
				case "WorkerServer", "MasterServer", "ApiApplicationServer", "AlertServer":
					dolphinscheduler = false
				case "QuorumPeerMain":
					zookeeper = false
				}
			}
		}

		// Add node resource data to resource status list
		HostResourceStatus = append(HostResourceStatus, ComponentStatus{
			Name: node.Name,
			CPU:  cpuUsage,
			RAM:  memUsage,
			ROM:  diskUsage,
		})
	}

	return hadoop, dorisFE, dorisBE, kafka, dolphinscheduler, zookeeper, HostResourceStatus
}
