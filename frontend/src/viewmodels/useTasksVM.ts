import { useMemo, useState } from 'react'
import { taskBoard } from '../mock/data'
import type { Task, TaskStatus, TaskBoardSummary } from '../models/types'

interface TaskView extends Task {
  status: TaskStatus
}

interface TaskGroupView {
  id: string
  name: string
  scoreEarned: number
  scoreMax: number
  isSpecial?: boolean
  deadlines: TaskBoardSummary['groups'][number]['deadlines']
  tasks: TaskView[]
}

export interface TasksVM {
  board: TaskBoardSummary
  groups: TaskGroupView[]
  showPastDeadlines: boolean
  togglePastDeadlines: () => void
}

const getTaskStatus = (task: Task): TaskStatus => {
  if (task.scoreEarned > task.score) {
    return 'over_solved'
  }
  if (task.scoreEarned === task.score) {
    return 'solved'
  }
  if (task.scoreEarned > 0) {
    return 'solved_partially'
  }
  return 'unsolved'
}

export function useTasksVM(): TasksVM {
  const [showPastDeadlines, setShowPastDeadlines] = useState(false)

  const groups = useMemo(() => {
    return taskBoard.groups.map((group) => {
      const tasks: TaskView[] = group.tasks.map((task) => ({
        ...task,
        status: getTaskStatus(task),
      }))

      const scoreMax = group.tasks
        .filter((task) => !task.isBonus)
        .reduce((total, task) => total + task.score, 0)
      const scoreEarned = group.tasks.reduce((total, task) => total + task.scoreEarned, 0)

      return {
        id: group.id,
        name: group.name,
        isSpecial: group.isSpecial,
        deadlines: group.deadlines,
        tasks,
        scoreMax,
        scoreEarned,
      }
    })
  }, [])

  return {
    board: taskBoard,
    groups,
    showPastDeadlines,
    togglePastDeadlines: () => setShowPastDeadlines((value) => !value),
  }
}
