import { useMemo, useState } from 'react'
import { courses } from '../mock/data'
import type { Course } from '../models/types'

export interface CoursesVM {
  activeCourses: Course[]
  finishedCourses: Course[]
  showFinished: boolean
  toggleFinished: () => void
}

export function useCoursesVM(): CoursesVM {
  const [showFinished, setShowFinished] = useState(false)

  const { activeCourses, finishedCourses } = useMemo(() => {
    const finished: Course[] = []
    const active: Course[] = []

    courses.forEach((course) => {
      if (course.status === 'finished') {
        finished.push(course)
      } else {
        active.push(course)
      }
    })

    return { activeCourses: active, finishedCourses: finished }
  }, [])

  return {
    activeCourses,
    finishedCourses,
    showFinished,
    toggleFinished: () => setShowFinished((value) => !value),
  }
}
