export type CourseStatus = 'created' | 'hidden' | 'in_progress' | 'all_tasks_issued' | 'doreshka' | 'finished'

export type TaskStatus = 'unsolved' | 'solved' | 'solved_partially' | 'over_solved'

export type DeadlineStatus = 'active' | 'urgent' | 'expired'

export interface Course {
  id: string
  name: string
  status: CourseStatus
  url: string
  isFinished?: boolean
}

export interface CourseSummary {
  name: string
  url: string
}

export interface UserProfile {
  username: string
  initials: string
  role: 'student' | 'namespace_admin' | 'program_manager' | 'instance_admin'
}

export interface Task {
  id: string
  name: string
  score: number
  scoreEarned: number
  stats: number
  isBonus?: boolean
  isSpecial?: boolean
  url?: string
}

export interface Deadline {
  id: string
  label: string
  percent: number
  dueAt: string
  status: DeadlineStatus
}

export interface TaskGroup {
  id: string
  name: string
  isSpecial?: boolean
  startedAt: string
  endsAt: string
  deadlines: Deadline[]
  tasks: Task[]
}

export interface TaskBoardSummary {
  courseName: string
  courseStatus: CourseStatus
  solvedScore: number
  maxScore: number
  solvedPercent: number
  groups: TaskGroup[]
}

export interface Namespace {
  id: string
  name: string
  slug: string
  description?: string
  gitlabGroupId: string
  coursesCount: number
  usersCount: number
}

export interface NamespaceUser {
  id: string
  username: string
  rmsId: string
  role: 'namespace_admin' | 'program_manager' | 'student'
}

export interface NamespaceCourse {
  id: string
  name: string
  status: 'running' | 'created' | 'hidden' | 'finished'
  gitlabGroup: string
  owners: string[]
  url: string
}

export interface InstancePanelSummary {
  totalCourses: number
  totalUsers: number
  totalNamespaces: number
  healthStatus: 'ok' | 'degraded'
}
