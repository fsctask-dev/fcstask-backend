import { Link } from 'react-router-dom'
import './Pages.css'

export function NotReadyPage() {
  return (
    <section className="page-grid">
      <div className="panel hero">
        <p className="eyebrow">Notice</p>
        <h1>Course is not ready yet</h1>
        <p className="subtle">
          We are still preparing the tasks, tests, and dashboards. Check back soon or contact your admin.
        </p>
        <div className="hero__actions">
          <Link className="btn" to="/">
            Back to courses
          </Link>
          <Link className="btn btn-ghost" to="/signup">
            Join with invite
          </Link>
        </div>
      </div>
    </section>
  )
}
