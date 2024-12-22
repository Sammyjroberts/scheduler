package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"turionspace/nei-mission-planner/scheduler/scheduler"
)

func main() {
	// Create a fixed start time for better readability
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

	tasks := []scheduler.Task{
		// Morning Tasks (9:00 - 12:00)
		{
			StartTime: baseTime,                    // 9:00
			EndTime:   baseTime.Add(3 * time.Hour), // 12:00
			Priority:  15.0,                        // Long high-priority meeting
		},
		{
			StartTime: baseTime,                    // 9:00
			EndTime:   baseTime.Add(1 * time.Hour), // 10:00
			Priority:  8.0,                         // Short morning task
		},
		{
			StartTime: baseTime.Add(30 * time.Minute), // 9:30
			EndTime:   baseTime.Add(90 * time.Minute), // 10:30
			Priority:  12.0,                           // Overlaps with multiple tasks
		},

		// Mid-Morning Tasks (10:00 - 11:00)
		{
			StartTime: baseTime.Add(1 * time.Hour), // 10:00
			EndTime:   baseTime.Add(2 * time.Hour), // 11:00
			Priority:  9.0,                         // Medium priority
		},
		{
			StartTime: baseTime.Add(75 * time.Minute),  // 10:15
			EndTime:   baseTime.Add(105 * time.Minute), // 10:45
			Priority:  7.0,                             // Short overlapping task
		},

		// Late Morning Tasks (11:00 - 13:00)
		{
			StartTime: baseTime.Add(2 * time.Hour), // 11:00
			EndTime:   baseTime.Add(4 * time.Hour), // 13:00
			Priority:  20.0,                        // Highest priority long task
		},
		{
			StartTime: baseTime.Add(150 * time.Minute), // 11:30
			EndTime:   baseTime.Add(180 * time.Minute), // 12:00
			Priority:  11.0,                            // Overlaps with high priority
		},

		// Afternoon Tasks (13:00 - 17:00)
		{
			StartTime: baseTime.Add(4 * time.Hour), // 13:00
			EndTime:   baseTime.Add(5 * time.Hour), // 14:00
			Priority:  6.0,                         // Lower priority
		},
		{
			StartTime: baseTime.Add(4*time.Hour + 30*time.Minute), // 13:30
			EndTime:   baseTime.Add(6 * time.Hour),                // 15:00
			Priority:  10.0,                                       // Medium-long task
		},
		{
			StartTime: baseTime.Add(5 * time.Hour), // 14:00
			EndTime:   baseTime.Add(7 * time.Hour), // 16:00
			Priority:  13.0,                        // Long afternoon task
		},
		{
			StartTime: baseTime.Add(6 * time.Hour), // 15:00
			EndTime:   baseTime.Add(8 * time.Hour), // 17:00
			Priority:  16.0,                        // High priority end of day
		},

		// Quick Tasks Throughout Day
		{
			StartTime: baseTime.Add(2*time.Hour + 30*time.Minute), // 11:30
			EndTime:   baseTime.Add(2*time.Hour + 45*time.Minute), // 11:45
			Priority:  5.0,                                        // Short task
		},
		{
			StartTime: baseTime.Add(5*time.Hour + 30*time.Minute), // 14:30
			EndTime:   baseTime.Add(5*time.Hour + 45*time.Minute), // 14:45
			Priority:  4.0,                                        // Quick afternoon task
		},

		// Zero Duration Tasks
		{
			StartTime: baseTime.Add(3 * time.Hour), // 12:00
			EndTime:   baseTime.Add(3 * time.Hour), // 12:00
			Priority:  3.0,                         // Instant task 1
		},
		{
			StartTime: baseTime.Add(3 * time.Hour), // 12:00
			EndTime:   baseTime.Add(3 * time.Hour), // 12:00
			Priority:  7.0,                         // Instant task 2 (same time)
		},
	}

	chosenTasks, totalPriority := scheduler.FindBestSchedule(tasks)

	// Print results in a nice format
	fmt.Println("\n🗓️  Optimal Schedule:")
	fmt.Println("------------------------------------------------")
	for _, task := range chosenTasks {
		fmt.Printf("   Start: %s\n", task.StartTime.Format("15:04"))
		fmt.Printf("   End: %s\n", task.EndTime.Format("15:04"))
		fmt.Printf("   Priority: %.1f\n", task.Priority)
		fmt.Println("------------------------------------------------")
	}
	fmt.Printf("\n📊 Total Priority Score: %.1f\n", totalPriority)

	// Print some statistics
	fmt.Printf("\n📈 Schedule Statistics:")
	fmt.Printf("\n   Total Tasks Available: %d", len(tasks))
	fmt.Printf("\n   Tasks Scheduled: %d", len(chosenTasks))
	fmt.Printf("\n   Time Span: %s - %s\n",
		baseTime.Format("15:04"),
		baseTime.Add(8*time.Hour).Format("15:04"))
	// Create sets for easy lookup of chosen tasks
	chosenMap := make(map[time.Time]bool)
	for _, task := range chosenTasks {
		chosenMap[task.StartTime] = true
	}

	// Prepare rejected tasks
	rejectedTasks := make([]scheduler.Task, 0)
	for _, task := range tasks {
		if !chosenMap[task.StartTime] {
			rejectedTasks = append(rejectedTasks, task)
		}
	}

	// Convert to output format
	chosenOutput := make([]scheduler.TaskOutput, len(chosenTasks))
	for i, task := range chosenTasks {
		chosenOutput[i] = scheduler.TaskOutput{
			StartTime:      task.StartTime.Format(time.RFC3339),
			EndTime:        task.EndTime.Format(time.RFC3339),
			Priority:       task.Priority,
			DurationMins:   int(task.EndTime.Sub(task.StartTime).Minutes()),
			IsZeroDuration: !task.EndTime.After(task.StartTime),
		}
	}

	rejectedOutput := make([]scheduler.TaskOutput, len(rejectedTasks))
	for i, task := range rejectedTasks {
		rejectedOutput[i] = scheduler.TaskOutput{
			StartTime:      task.StartTime.Format(time.RFC3339),
			EndTime:        task.EndTime.Format(time.RFC3339),
			Priority:       task.Priority,
			DurationMins:   int(task.EndTime.Sub(task.StartTime).Minutes()),
			IsZeroDuration: !task.EndTime.After(task.StartTime),
		}
	}
	// Create final output structure
	output := scheduler.ScheduleOutput{
		ChosenTasks:   chosenOutput,
		RejectedTasks: rejectedOutput,
		TotalPriority: totalPriority,
		Statistics: scheduler.Statistics{
			TotalTasks:     len(tasks),
			ScheduledTasks: len(chosenTasks),
			RejectedTasks:  len(rejectedTasks),
		},
		TimeRange: scheduler.TimeRange{
			Start: baseTime.Format(time.RFC3339),
			End:   baseTime.Add(8 * time.Hour).Format(time.RFC3339),
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Print JSON
	fmt.Println(string(jsonData))
	// save to file
	err = os.WriteFile("output.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %v\n", err)
		return
	}
}
