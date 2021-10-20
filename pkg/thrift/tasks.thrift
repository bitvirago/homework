namespace go task

struct TaskResponse {
  1: string id
  2: string command
}

struct TaskUpdateRequest {
  1: string id
  2: string started_at
  3: string finished_at
  4: string status
  5: string stdout
  6: string stderr
  7: i32    exit_code
}

service TaskService {
  TaskResponse getNextTask()
  void updateTask(1:TaskUpdateRequest updateRequest)
}