Use Warden before any finish claim.

1. Call `validate_plan` if recent plan edits happened.
2. Call `request_finish`.
3. If `can_finish` is false, do not stop. Explain the blocking reasons and continue with the recommended next task.
4. Only report completion if Warden approves finish.