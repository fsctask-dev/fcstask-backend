import { Link } from 'react-router-dom'
import { useCoursesVM } from '../viewmodels/useCoursesVM'
import './Pages.css'

export function CoursesPage() {
  const { activeCourses, finishedCourses, showFinished, toggleFinished } = useCoursesVM()
  const courseMeta: Record<string, { progress: number; next: string }> = {
    algorithms: { progress: 63, next: 'Next deadline in 3 days' },
    mlops: { progress: 82, next: 'Next deadline in 8 days' },
    rust: { progress: 12, next: 'Schedule starts soon' },
    golang: { progress: 100, next: 'Course finished' },
    'advanced-cpp': { progress: 48, next: 'Next deadline in 5 days' },
    'advanced-python': { progress: 0, next: 'Schedule starts soon' },
  }

  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Overview</p>
          <h1>Courses</h1>
          <p className="subtle">Keep track of your programs, tasks, and results.</p>
        </div>
        <div className="header-actions">
          <button className="btn btn-ghost" onClick={toggleFinished} type="button">
            {showFinished ? 'Hide completed' : 'Show completed'}
          </button>
          <Link className="btn" to="/course/create">
            Create course
          </Link>
        </div>
      </div>

      {showFinished && (
        <div className="panel">
          <h2>Completed courses</h2>
          <div className="course-grid">
            {finishedCourses.length === 0 ? (
              <p className="empty">No finished courses yet.</p>
            ) : (
              finishedCourses.map((course) => (
                <Link key={course.id} to={course.url} className="course-card course-card--complete">
                  <div className="course-card__top">
                    <div>
                      <p className="course-card__eyebrow">Completed</p>
                      <h3>{course.name}</h3>
                    </div>
                    <span className="status status--finished">{course.status.replace('_', ' ')}</span>
                  </div>
                  <div className="course-card__footer">
                    <span className="meta">Results archived</span>
                    <span className="course-card__badge">100%</span>
                  </div>
                </Link>
              ))
            )}
          </div>
        </div>
      )}

      <div className="panel">
        <h2>Active courses</h2>
        <div className="course-grid">
          {activeCourses.map((course) => (
            <Link key={course.id} to={course.url} className="course-card">
              <div className="course-card__top">
                <div>
                  <p className="course-card__eyebrow">In progress</p>
                  <h3>{course.name}</h3>
                  <p className="meta">{courseMeta[course.id]?.next ?? 'Next deadline soon'}</p>
                </div>
                <span className={`status status--${course.status}`}>{course.status.replace('_', ' ')}</span>
              </div>
              <div className="course-card__progress">
                <div className="course-card__progress-bar">
                  <span style={{ width: `${courseMeta[course.id]?.progress ?? 0}%` }} />
                </div>
                <span className="course-card__badge">{courseMeta[course.id]?.progress ?? 0}%</span>
              </div>
            </Link>
          ))}
        </div>
      </div>

      <div className="panel">
        <h2>Register for a new course</h2>
        <form className="form-inline" onSubmit={(event) => event.preventDefault()}>
          <input className="input" placeholder="Course title..." />
          <button className="btn btn-ghost" type="submit">
            Go
          </button>
        </form>
      </div>
    </section>
  )
}
