package schedules

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"joodpakpaw/app/scheduler"
	"joodpakpaw/config"
	"joodpakpaw/internal/helpers"
	"joodpakpaw/internal/helpers/color"
)

type (
	SchedulerTask struct {
		sc          scheduler.IScheduler
		cronString  string
		mailService *helpers.MailService
	}
)

func NewSchedulerTask(
	cronString string,
) *SchedulerTask {
	mailConfig := config.NewMailConfig()
	return &SchedulerTask{
		sc:          scheduler.CurrentScheduler(),
		cronString:  cronString,
		mailService: helpers.NewMailService(mailConfig.SmtpHost, mailConfig.SmtpPort, mailConfig.SmtpUsername, mailConfig.SmtpPassword),
	}
}

// StartScheduler initializes the global scheduler and registers all scheduled tasks.
// ! This function should be called in the main application to register all scheduled tasks.
func StartScheduler() {
	if scheduler.GlobalScheduler == nil {
		return
	}

	// Register all jobs
	RegisterJobs()

	// Print schedule task count
	if !fiber.IsChild() {
		log.Println("ScheduleTask:", color.Format(color.GREEN, fmt.Sprint(scheduler.GlobalScheduler.GetTaskCount())), color.Format(color.GREEN, "jobs"))
		// Print break line
		log.Println("--------------------------------------------------")
	}

	// Start scheduler service to run all jobs in the background (non-blocking)
	go scheduler.GlobalScheduler.Start()
}

// RegisterJobs registers all the jobs to the scheduler. (Cron format: minute, hour, day, month, day of week)
func RegisterJobs() {
	// Add scheduler job here... (minute, hour, day, month, day of week)
	NewSchedulerTask("* * * * *").ExampleScheduleJob()
	NewSchedulerTask("59 23 * * *").BookingAutoSuccessJob()
	NewSchedulerTask("0 * * * *").BookingAutoCancelJob()
	NewSchedulerTask("0 8 * * *").BookingReminderJob()
}
