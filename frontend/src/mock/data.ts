import type {
  Course,
  CourseSummary,
  InstancePanelSummary,
  Namespace,
  NamespaceCourse,
  NamespaceUser,
  TaskBoardSummary,
  UserProfile,
} from '../models/types'

export const currentUser: UserProfile = {
  username: 'student',
  initials: 'ST',
  role: 'instance_admin',
}

export const courses: Course[] = [
  {
    id: 'algorithms',
    name: 'Algorithms 101',
    status: 'in_progress',
    url: '/course/algorithms',
  },
  {
    id: 'mlops',
    name: 'MLOps Studio',
    status: 'all_tasks_issued',
    url: '/course/mlops',
  },
  {
    id: 'rust',
    name: 'Rust Core',
    status: 'created',
    url: '/course/rust',
  },
  {
    id: 'golang',
    name: 'Go Lab',
    status: 'finished',
    url: '/course/golang',
  },
  {
    id: 'advanced-cpp',
    name: 'Advanced C++',
    status: 'in_progress',
    url: '/course/advanced-cpp',
  },
  {
    id: 'advanced-python',
    name: 'Advanced Python',
    status: 'created',
    url: '/course/advanced-python',
  },
]

export const coursesShort: CourseSummary[] = courses.map((course) => ({
  name: course.name,
  url: course.url,
}))

export const taskBoard: TaskBoardSummary = {
  courseName: 'Algorithms 101',
  courseStatus: 'in_progress',
  solvedScore: 126,
  maxScore: 200,
  solvedPercent: 63,
  groups: [
    {
      id: 'week-1',
      name: 'Week 1: Warmup',
      startedAt: '2024-10-01T09:00:00Z',
      endsAt: '2024-10-14T18:00:00Z',
      deadlines: [
        {
          id: 'd1',
          label: 'Checkpoint',
          percent: 0.6,
          dueAt: '2024-09-20T18:00:00Z',
          status: 'expired',
        },
        {
          id: 'd2',
          label: 'Final',
          percent: 1,
          dueAt: '2024-10-14T18:00:00Z',
          status: 'urgent',
        },
      ],
      tasks: [
        {
          id: 't1',
          name: 'Arrays Sprint',
          score: 20,
          scoreEarned: 20,
          stats: 0.82,
        },
        {
          id: 't2',
          name: 'Stack Trace',
          score: 25,
          scoreEarned: 10,
          stats: 0.64,
        },
        {
          id: 't3',
          name: 'Sorting Arena',
          score: 30,
          scoreEarned: 0,
          stats: 0.38,
          isSpecial: true,
        },
      ],
    },
    {
      id: 'week-2',
      name: 'Week 2: Graphs',
      startedAt: '2024-10-15T09:00:00Z',
      endsAt: '2024-10-28T18:00:00Z',
      isSpecial: true,
      deadlines: [
        {
          id: 'd3',
          label: 'Checkpoint',
          percent: 0.5,
          dueAt: '2024-10-22T18:00:00Z',
          status: 'active',
        },
        {
          id: 'd4',
          label: 'Final',
          percent: 1,
          dueAt: '2024-10-28T18:00:00Z',
          status: 'active',
        },
      ],
      tasks: [
        {
          id: 't4',
          name: 'Bridge Builder',
          score: 40,
          scoreEarned: 25,
          stats: 0.57,
        },
        {
          id: 't5',
          name: 'Shortest Path Lab',
          score: 30,
          scoreEarned: 0,
          stats: 0.44,
        },
        {
          id: 't6',
          name: 'Bonus Relay',
          score: 10,
          scoreEarned: 12,
          stats: 0.91,
          isBonus: true,
        },
      ],
    },
  ],
}

export const namespaces: Namespace[] = [
  {
    id: 'ns-01',
    name: 'Core CS',
    slug: 'core-cs',
    description: 'Foundational tracks for new cohorts.',
    gitlabGroupId: '22411',
    coursesCount: 5,
    usersCount: 180,
  },
  {
    id: 'ns-02',
    name: 'Applied ML',
    slug: 'applied-ml',
    description: 'Production-ready ML tasks and MLOps labs.',
    gitlabGroupId: '23898',
    coursesCount: 3,
    usersCount: 96,
  },
]

export const namespaceUsers: NamespaceUser[] = [
  {
    id: 'u-1',
    username: 'alex',
    rmsId: 'rms-210',
    role: 'namespace_admin',
  },
  {
    id: 'u-2',
    username: 'maria',
    rmsId: 'rms-218',
    role: 'program_manager',
  },
  {
    id: 'u-3',
    username: 'sasha',
    rmsId: 'rms-228',
    role: 'student',
  },
]

export const namespaceCourses: NamespaceCourse[] = [
  {
    id: 'c-101',
    name: 'Algorithms 101',
    status: 'running',
    gitlabGroup: 'algorithms-101',
    owners: ['alex', 'maria'],
    url: '/course/algorithms',
  },
  {
    id: 'c-102',
    name: 'Rust Core',
    status: 'created',
    gitlabGroup: 'rust-core',
    owners: ['alex'],
    url: '/course/rust',
  },
]

export const instancePanel: InstancePanelSummary = {
  totalCourses: 21,
  totalUsers: 920,
  totalNamespaces: 6,
  healthStatus: 'ok',
}
