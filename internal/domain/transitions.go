package domain

var taskTransitions = map[TaskStatus]map[TaskStatus]struct{}{
	TaskNotStarted: {TaskInProgress: {}, TaskBlocked: {}, TaskCancelled: {}},
	TaskInProgress: {TaskDone: {}, TaskBlocked: {}, TaskNotStarted: {}},
	TaskBlocked:    {TaskInProgress: {}, TaskWaived: {}, TaskCancelled: {}},
	TaskDone:       {},
	TaskCancelled:  {},
	TaskWaived:     {},
}

var phaseTransitions = map[PhaseStatus]map[PhaseStatus]struct{}{
	PhaseNotStarted: {PhaseInProgress: {}, PhaseBlocked: {}},
	PhaseInProgress: {PhaseBlocked: {}, PhaseCompleted: {}, PhaseNotStarted: {}},
	PhaseBlocked:    {PhaseInProgress: {}},
	PhaseCompleted:  {},
}

var planTransitions = map[PlanStatus]map[PlanStatus]struct{}{
	PlanDraft:     {PlanActive: {}, PlanArchived: {}},
	PlanActive:    {PlanBlocked: {}, PlanCompleted: {}, PlanArchived: {}},
	PlanBlocked:   {PlanActive: {}, PlanArchived: {}},
	PlanCompleted: {PlanArchived: {}},
	PlanArchived:  {},
}

func CanTransitionTask(current, target TaskStatus) bool {
	_, ok := taskTransitions[current][target]
	return ok
}

func CanTransitionPhase(current, target PhaseStatus) bool {
	_, ok := phaseTransitions[current][target]
	return ok
}

func CanTransitionPlan(current, target PlanStatus) bool {
	_, ok := planTransitions[current][target]
	return ok
}
