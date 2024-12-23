package scheduler

import (
	"context"
	"sort"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SchedulerConfig struct {
	fx.In
	Logger *otelzap.Logger
}

func NewScheduler(cfg SchedulerConfig) *Scheduler {
	return &Scheduler{
		logger: cfg.Logger,
	}
}

type Scheduler struct {
	logger *otelzap.Logger
}

// isZeroDuration checks if a task has zero duration
func (s *Scheduler) isZeroDuration(task Task) bool {
	return !task.EndTime.After(task.StartTime)
}

// tasksConflict checks if two tasks overlap, treating zero duration tasks as regular tasks
func (s *Scheduler) tasksConflict(task1, task2 Task) bool {
	// For zero duration tasks, they conflict if they happen at the same instant
	if s.isZeroDuration(task1) && s.isZeroDuration(task2) {
		return task1.StartTime.Equal(task2.StartTime)
	}

	// If one task is zero duration, it conflicts if it occurs during the other task
	if s.isZeroDuration(task1) {
		return !task1.StartTime.Before(task2.StartTime) && !task1.StartTime.After(task2.EndTime)
	}
	if s.isZeroDuration(task2) {
		return !task2.StartTime.Before(task1.StartTime) && !task2.StartTime.After(task1.EndTime)
	}

	// Regular overlap check for non-zero duration tasks
	return task1.StartTime.Before(task2.EndTime) && task2.StartTime.Before(task1.EndTime)
}

// findBestPreviousTask finds the most recent task that doesn't overlap with our current task
// this should work with our dp solution because (need shri to check this)
// We've sorted by end time
// When we use this result, we then look up that task's accumulated best priority in bestPriorityUpToTask
// The dynamic programming table has already calculated the optimal priority for all previous positions
func (s *Scheduler) findBestPreviousTask(tasks []Task, currentTaskIndex int) int {
	currentTask := tasks[currentTaskIndex]

	// Binary search through previous tasks
	startSearch := 0
	endSearch := currentTaskIndex - 1
	bestPreviousTask := -1

	for startSearch <= endSearch {
		middleTask := (startSearch + endSearch) / 2

		// Check for conflict
		if !s.tasksConflict(tasks[middleTask], currentTask) {
			bestPreviousTask = middleTask
			startSearch = middleTask + 1 // Look for an even later task
		} else {
			endSearch = middleTask - 1 // This task overlaps, look earlier
		}
	}

	return bestPreviousTask
}

// FindBestSchedule finds the combination of tasks that gives us the highest total priority
func (s *Scheduler) FindBestSchedule(tasks []Task) ([]Task, float64, []RejectedTask) {
	ctx, span := otel.GetTracerProvider().Tracer("scheduler").Start(context.Background(), "FindBestSchedule")
	defer span.End()
	logger := s.logger.Ctx(ctx)
	span.SetAttributes(attribute.Int("num_tasks", len(tasks)))
	logger.Info("Starting scheduler", zap.Int("num_tasks", len(tasks)))
	rejectedTasks := []RejectedTask{}
	// if there are no tasks, return nil
	if len(tasks) == 0 {
		return nil, 0, nil
	}

	// Sort tasks by end time - zero duration tasks are sorted by their start time
	sort.Slice(tasks, func(first, second int) bool {
		// For zero duration tasks, use their start time
		firstTime := tasks[first].EndTime
		secondTime := tasks[second].EndTime
		if s.isZeroDuration(tasks[first]) {
			firstTime = tasks[first].StartTime
		}
		if s.isZeroDuration(tasks[second]) {
			secondTime = tasks[second].StartTime
		}
		return firstTime.Before(secondTime)
	})
	// Initialize our dynamic programming arrays
	numTasks := len(tasks)
	// bestPriorityUpToTask stores the best priority we can get up to a given task
	bestPriorityUpToTask := make([]float64, numTasks)
	// previousTaskChosen stores the index of the task that was chosen before the current task
	previousTaskChosen := make([]int, numTasks)

	// Base case
	bestPriorityUpToTask[0] = tasks[0].Priority
	previousTaskChosen[0] = -1

	// For each task, figure out the best way to include it
	for currentTask := 1; currentTask < numTasks; currentTask++ {
		// Find the index of the latest task that finishes before the current task starts
		// and does not overlap with it. This is the best candidate to have been
		// included in the schedule *before* the current task.
		bestPrevious := s.findBestPreviousTask(tasks, currentTask)

		// Calculate the total priority if we *include* the current task.
		priorityIfIncluded := tasks[currentTask].Priority
		// If there's a compatible previous task, add its accumulated priority to the
		// priority we get by including the current task. This is the core of the DP logic:
		// we're reusing previously calculated optimal solutions for subproblems.
		if bestPrevious != -1 {
			priorityIfIncluded += bestPriorityUpToTask[bestPrevious]
		}

		// Calculate the total priority if we *exclude* the current task.
		// In this case, the best priority we can achieve is simply the best priority
		// we could achieve up to the *previous* task (currentTask - 1).
		priorityIfExcluded := bestPriorityUpToTask[currentTask-1]

		// Now, we make the optimal choice: do we include the current task or not?
		if priorityIfIncluded > priorityIfExcluded {
			// Including the current task gives us a higher total priority.
			// So, we update the bestPriorityUpToTask for the current task to reflect this.
			bestPriorityUpToTask[currentTask] = priorityIfIncluded
			// We also record the index of the previous task that was part of this optimal
			// solution. This is crucial for reconstructing the actual schedule later.
			previousTaskChosen[currentTask] = bestPrevious
		} else {
			// Excluding the current task gives us a higher or equal total priority.
			// We keep the best priority we had up to the previous task.
			bestPriorityUpToTask[currentTask] = priorityIfExcluded
			// If we exclude the current task, the previous task chosen remains the same
			// as the one chosen for the previous iteration. This maintains the chain
			// of chosen tasks for backtracking.
			previousTaskChosen[currentTask] = previousTaskChosen[currentTask-1]
			// Record low priority rejection
			span.AddEvent("task_rejected", trace.WithAttributes(attribute.String("reason", "low_priority")))
			rejectedTasks = append(rejectedTasks, RejectedTask{
				TaskRejected: tasks[currentTask],
				Reason:       RejectionReasonLowPriority,
			})
		}
	}

	// Build our list of chosen tasks
	chosenTasks := make([]Task, 0)
	// Build chosen tasks list and track conflicts
	chosenIndexes := make(map[int]bool)

	for i := numTasks - 1; i >= 0; {
		if i == 0 || bestPriorityUpToTask[i] != bestPriorityUpToTask[i-1] {
			chosenTasks = append(chosenTasks, tasks[i])
			chosenIndexes[i] = true
			i = previousTaskChosen[i]
		} else {
			i--
		}
	}

	// Find conflict rejections
	for i := 0; i < numTasks; i++ {
		if !chosenIndexes[i] {
			// Check if already rejected for low priority
			alreadyRejected := false
			for _, rejected := range rejectedTasks {
				if rejected.TaskRejected == tasks[i] {
					alreadyRejected = true
					break
				}
			}

			if !alreadyRejected {
				// Find conflicting task
				for j := 0; j < numTasks; j++ {
					if chosenIndexes[j] && s.tasksConflict(tasks[i], tasks[j]) {
						span.AddEvent("task_rejected", trace.WithAttributes(attribute.String("reason", "conflict")))
						rejectedTasks = append(rejectedTasks, RejectedTask{
							TaskRejected: tasks[i],
							CausedBy:     &tasks[j],
							Reason:       RejectionReasonConflict,
						})
						break
					}
				}
			}
		}
	}
	for i := numTasks - 1; i >= 0; {
		if i == 0 || bestPriorityUpToTask[i] != bestPriorityUpToTask[i-1] {
			chosenTasks = append(chosenTasks, tasks[i])
			i = previousTaskChosen[i]
		} else {
			i--
		}
	}

	// Put tasks in chronological order
	for i := 0; i < len(chosenTasks)/2; i++ {
		chosenTasks[i], chosenTasks[len(chosenTasks)-1-i] = chosenTasks[len(chosenTasks)-1-i], chosenTasks[i]
	}
	span.AddEvent("scheduler_finished", trace.WithAttributes(attribute.Int("num_chosen_tasks", len(chosenTasks)), attribute.Int("num_rejected_tasks", len(rejectedTasks))))
	logger.Info("Scheduler finished", zap.Int("num_chosen_tasks", len(chosenTasks)), zap.Int("num_rejected_tasks", len(rejectedTasks)))
	return chosenTasks, bestPriorityUpToTask[numTasks-1], rejectedTasks
}

var Module = fx.Provide(NewScheduler)
