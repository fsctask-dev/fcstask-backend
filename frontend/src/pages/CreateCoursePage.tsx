import { CourseForm } from '../components/CourseForm'
import { useCourseFormVM } from '../viewmodels/useCourseFormVM'
import './Pages.css'

export function CreateCoursePage() {
  const { form, updateField } = useCourseFormVM('create')

  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Admin</p>
          <h1>Create course</h1>
          <p className="subtle">Set up a new learning track and its repository.</p>
        </div>
      </div>

      <div className="panel">
        <CourseForm form={form} onChange={updateField} submitLabel="Create course" />
      </div>
    </section>
  )
}
