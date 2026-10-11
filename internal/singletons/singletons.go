package singletons

import (
	"log"

	"github.com/terabiome/terabiome/internal/config"
	"github.com/terabiome/terabiome/internal/services"
	"github.com/terabiome/terabiome/pkg/logging"
)

type Singletons struct {
	workflowService *services.WorkflowService
}

var sings *Singletons

func Init() {
	log.Println("Constructing singletons")

	logging.InitGlobalSlogger(config.Get().Logger)

	workflowService := services.NewWorkflowService()
	sings = &Singletons{
		workflowService: workflowService,
	}
	// showing off

	logging.GetGlobalSlogger().Info("Constructed singletons")
}

func GetWorkflowService() *services.WorkflowService {
	return sings.workflowService
}
