import { useState } from 'react'

export interface CourseFormState {
  name: string
  slug: string
  status: 'created' | 'hidden' | 'in_progress' | 'finished'
  startDate: string
  endDate: string
  repoTemplate: string
  description: string
}

export function useCourseFormVM(mode: 'create' | 'edit') {
  const [form, setForm] = useState<CourseFormState>({
    name: mode === 'create' ? '' : 'Algorithms 101',
    slug: mode === 'create' ? '' : 'algorithms',
    status: mode === 'create' ? 'created' : 'in_progress',
    startDate: '2024-10-01',
    endDate: '2024-12-20',
    repoTemplate: mode === 'create' ? '' : 'git@gitlab.local/course-template.git',
    description: 'Core practice track for the semester.',
  })

  const updateField = <K extends keyof CourseFormState>(key: K, value: CourseFormState[K]) => {
    setForm((current) => ({
      ...current,
      [key]: value,
    }))
  }

  return {
    form,
    updateField,
  }
}
