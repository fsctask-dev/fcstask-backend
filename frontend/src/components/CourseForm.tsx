import type { CourseFormState } from '../viewmodels/useCourseFormVM'
import './CourseForm.css'

interface CourseFormProps {
  form: CourseFormState
  onChange: <K extends keyof CourseFormState>(key: K, value: CourseFormState[K]) => void
  submitLabel: string
}

export function CourseForm({ form, onChange, submitLabel }: CourseFormProps) {
  return (
    <form className="course-form" onSubmit={(event) => event.preventDefault()}>
      <label>
        Course name
        <input
          className="input"
          value={form.name}
          onChange={(event) => onChange('name', event.target.value)}
        />
      </label>
      <label>
        Slug
        <input
          className="input"
          value={form.slug}
          onChange={(event) => onChange('slug', event.target.value)}
        />
      </label>
      <label>
        Status
        <select
          className="input"
          value={form.status}
          onChange={(event) => onChange('status', event.target.value as CourseFormState['status'])}
        >
          <option value="created">Created</option>
          <option value="hidden">Hidden</option>
          <option value="in_progress">In progress</option>
          <option value="finished">Finished</option>
        </select>
      </label>
      <label>
        Start date
        <input
          className="input"
          type="date"
          value={form.startDate}
          onChange={(event) => onChange('startDate', event.target.value)}
        />
      </label>
      <label>
        End date
        <input
          className="input"
          type="date"
          value={form.endDate}
          onChange={(event) => onChange('endDate', event.target.value)}
        />
      </label>
      <label>
        Repo template
        <input
          className="input"
          value={form.repoTemplate}
          onChange={(event) => onChange('repoTemplate', event.target.value)}
        />
      </label>
      <label className="course-form__full">
        Description
        <textarea
          className="input"
          rows={4}
          value={form.description}
          onChange={(event) => onChange('description', event.target.value)}
        />
      </label>
      <div className="course-form__actions">
        <button className="btn" type="submit">
          {submitLabel}
        </button>
        <button className="btn btn-ghost" type="button">
          Cancel
        </button>
      </div>
    </form>
  )
}
