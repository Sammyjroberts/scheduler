package scheduler

import (
	"testing"
	"time"
)

// Helper function to create a fixed time for testing
func fixedTime(hour int) time.Time {
	return time.Date(2024, 1, 1, hour, 0, 0, 0, time.UTC)
}

// Helper function to compare two task slices
func tasksEqual(t *testing.T, expected, actual []Task) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("Length mismatch: expected %d tasks, got %d tasks", len(expected), len(actual))
		return
	}
	for i := range expected {
		if !expected[i].StartTime.Equal(actual[i].StartTime) {
			t.Errorf("Task %d start time mismatch: expected %v, got %v", i, expected[i].StartTime, actual[i].StartTime)
		}
		if !expected[i].EndTime.Equal(actual[i].EndTime) {
			t.Errorf("Task %d end time mismatch: expected %v, got %v", i, expected[i].EndTime, actual[i].EndTime)
		}
		if expected[i].Priority != actual[i].Priority {
			t.Errorf("Task %d priority mismatch: expected %.2f, got %.2f", i, expected[i].Priority, actual[i].Priority)
		}
	}
}

func TestFindBestPreviousTask(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []Task
		currentIndex  int
		expectedIndex int
	}{
		{
			name: "No previous compatible task",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 1},
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 2},
			},
			currentIndex:  1,
			expectedIndex: -1,
		},
		{
			name: "One compatible previous task",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 1},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 2},
			},
			currentIndex:  1,
			expectedIndex: 0,
		},
		{
			name: "Multiple compatible tasks, should find last one",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 1},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 2},
				{StartTime: fixedTime(11), EndTime: fixedTime(12), Priority: 3},
			},
			currentIndex:  2,
			expectedIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findBestPreviousTask(tt.tasks, tt.currentIndex)
			if result != tt.expectedIndex {
				t.Errorf("Expected index %d, got %d", tt.expectedIndex, result)
			}
		})
	}
}

func TestFindBestSchedule(t *testing.T) {
	tests := []struct {
		name             string
		tasks            []Task
		expectedTasks    []Task
		expectedPriority float64
	}{
		{
			name:             "Empty task list",
			tasks:            []Task{},
			expectedTasks:    []Task{},
			expectedPriority: 0,
		},
		{
			name: "Single task",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
			},
			expectedPriority: 5,
		},
		{
			name: "Two non-overlapping tasks",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 3},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 3},
			},
			expectedPriority: 8,
		},
		{
			name: "Two overlapping tasks - should pick higher priority",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 7},
				{StartTime: fixedTime(10), EndTime: fixedTime(12), Priority: 4},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 7},
			},
			expectedPriority: 7,
		},
		{
			name: "Three tasks - two short better than one long",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 10},
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 8},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 8},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 8},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 8},
			},
			expectedPriority: 16,
		},
		{
			name: "Complex overlapping scenario",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(12), Priority: 15}, // Long task
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 6},  // Short early
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 6}, // Short middle
				{StartTime: fixedTime(11), EndTime: fixedTime(12), Priority: 6}, // Short late
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 6},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 6},
				{StartTime: fixedTime(11), EndTime: fixedTime(12), Priority: 6},
			},
			expectedPriority: 18,
		},
		{
			name: "Tasks with equal end times",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 8},
				{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 5},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 8},
			},
			expectedPriority: 8,
		},
		{
			name: "Tasks with equal start times",
			tasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 8},
				{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
			},
			expectedTasks: []Task{
				{StartTime: fixedTime(9), EndTime: fixedTime(11), Priority: 8},
			},
			expectedPriority: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultTasks, resultPriority := FindBestSchedule(tt.tasks)

			if resultPriority != tt.expectedPriority {
				t.Errorf("Priority mismatch: expected %.2f, got %.2f", tt.expectedPriority, resultPriority)
			}

			tasksEqual(t, tt.expectedTasks, resultTasks)
		})
	}
}

// Test edge cases specifically
func TestEdgeCases(t *testing.T) {
	t.Run("Zero duration tasks", func(t *testing.T) {
		tasks := []Task{
			{StartTime: fixedTime(9), EndTime: fixedTime(9), Priority: 5},
			{StartTime: fixedTime(9), EndTime: fixedTime(9), Priority: 3},
		}
		resultTasks, resultPriority := FindBestSchedule(tasks)
		if len(resultTasks) != 1 {
			t.Errorf("Expected 1 task, got %d tasks", len(resultTasks))
		}
		if resultPriority != 5 {
			t.Errorf("Expected priority 5, got %.2f", resultPriority)
		}
	})

	t.Run("Negative duration tasks should still work", func(t *testing.T) {
		tasks := []Task{
			{StartTime: fixedTime(10), EndTime: fixedTime(9), Priority: 5},
		}
		resultTasks, _ := FindBestSchedule(tasks)
		if len(resultTasks) != 1 {
			t.Errorf("Expected 1 task, got %d tasks", len(resultTasks))
		}
	})

	t.Run("All same priority", func(t *testing.T) {
		tasks := []Task{
			{StartTime: fixedTime(9), EndTime: fixedTime(10), Priority: 5},
			{StartTime: fixedTime(10), EndTime: fixedTime(11), Priority: 5},
			{StartTime: fixedTime(11), EndTime: fixedTime(12), Priority: 5},
		}
		resultTasks, resultPriority := FindBestSchedule(tasks)
		if resultPriority != 15 {
			t.Errorf("Expected priority 15, got %.2f", resultPriority)
		}
		if len(resultTasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d tasks", len(resultTasks))
		}
	})
}

// Benchmark tests
func BenchmarkFindBestSchedule(b *testing.B) {
	// Create a large set of tasks for benchmarking
	tasks := make([]Task, 1000)
	for i := 0; i < 1000; i++ {
		tasks[i] = Task{
			StartTime: fixedTime(i),
			EndTime:   fixedTime(i + 2),
			Priority:  float64(i % 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindBestSchedule(tasks)
	}
}
