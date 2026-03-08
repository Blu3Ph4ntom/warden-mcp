Use Warden to decide the next required action.

1. Call `get_status`.
2. Call `get_next_task`.
3. If blocked, explain the blocking reasons and the smallest safe next step.
4. If not blocked, continue with the returned task instead of picking a task by intuition.