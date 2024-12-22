# Optimal Task Scheduler

This scheduler finds the optimal set of non-overlapping tasks that maximizes
total priority(for turion this will be a $ amount determined by customer /
custom org priority / quality of image). It uses dynamic programming with binary
search to solve the weighted interval scheduling problem.

## Problem Description

Given a set of tasks where each task has:

- Start time
- End time
- Priority (weight/value)

Find the subset of non-overlapping tasks that gives the highest total priority.

### Example

```
Input Tasks:
A: 9:00-10:00, Priority 5
B: 9:30-10:30, Priority 8
C: 10:00-11:00, Priority 4
D: 10:30-11:30, Priority 7

Optimal Schedule:
B (9:30-10:30, Priority 8)
D (10:30-11:30, Priority 7)
Total Priority: 15
```

## Algorithm Approach

The solution uses dynamic programming with three key insights:

1. Sort tasks by end time (crucial for the binary search to work)
2. For each task, we only need to know the best non-conflicting previous task
3. We can find the best previous task efficiently using binary search

### Dynamic Programming Solution

For each task, we solve two subproblems:

1. What's the best schedule including this task?
2. What's the best schedule excluding this task?

The recurrence relation is:

```
bestPriority[i] = max(
    task[i].priority + bestPriority[bestPrevious[i]],  // Include task
    bestPriority[i-1]                                  // Exclude task
)
```

### Example Step-by-Step

Given these tasks (sorted by end time):

```
A: 9:00-10:00, Priority 5
B: 9:30-10:30, Priority 8
C: 10:00-11:00, Priority 4
D: 10:30-11:30, Priority 7
```

The algorithm proceeds:

1. For task A:
   - First task, just take its priority
   - bestPriority[0] = 5

2. For task B:
   - Can't combine with A (overlap)
   - Choose max(B alone, A alone)
   - bestPriority[1] = max(8, 5) = 8

3. For task C:
   - Can combine with A
   - Choose max(C+A, keep B)
   - bestPriority[2] = max(4+5, 8) = 9

4. For task D:
   - Can combine with B
   - Choose max(D+B, keep C+A)
   - bestPriority[3] = max(7+8, 9) = 15

### Binary Search Optimization

Finding the best previous non-conflicting task uses binary search because:

1. Tasks are sorted by end time
2. If task X conflicts, all later tasks also conflict
3. If task X doesn't conflict, it's a candidate but there might be better ones
   later

Example binary search for task D (10:30-11:30):

```
1. Check middle task B (9:30-10:30):
   - No conflict
   - Save as candidate
   - Look for better tasks after B

2. Check task C (10:00-11:00):
   - Conflicts
   - Keep previous candidate (B)
   - Look for tasks before C

3. Search complete:
   - Best previous task is B
```

## Time Complexity

- Overall: O(n log n)
  - Sorting: O(n log n)
  - Dynamic Programming: O(n)
  - Binary Search for each task: O(log n) per task = O(n log n) total

## Space Complexity

- O(n) for storing:
  - Best priorities array
  - Previous task choices array
  - Final schedule

## Special Cases

The implementation handles:

- Zero duration tasks
- Tasks with equal start/end times
- Empty task lists
- Overlapping high priority vs multiple lower priority tasks

## Usage

```go
tasks := []Task{
    {
        StartTime: time.Now(),
        EndTime:   time.Now().Add(1 * time.Hour),
        Priority:  5.0,
    },
    // Add more tasks...
}

chosenTasks, totalPriority := FindBestSchedule(tasks)
```

## Visualization

The repository includes an HTML visualizer that shows:

- Optimal schedule
- Rejected tasks
- Task priorities and timings
- Total priority achieved

## Further Reading

- [Weighted Interval Scheduling](https://en.wikipedia.org/wiki/Interval_scheduling#Weighted)
- [Dynamic Programming](https://en.wikipedia.org/wiki/Dynamic_programming)
- [Binary Search](https://en.wikipedia.org/wiki/Binary_search_algorithm)
