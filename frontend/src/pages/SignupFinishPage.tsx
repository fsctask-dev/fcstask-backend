import { Link } from 'react-router-dom'
import './Pages.css'

export function SignupFinishPage() {
  return (
    <section className="auth-card">
      <div className="auth-card__header">
        <p className="eyebrow">All set</p>
        <h1>Welcome aboard</h1>
        <p className="subtle">Your registration is complete. The course will appear shortly.</p>
      </div>
      <div className="auth-actions">
        <Link className="btn" to="/">
          Go to dashboard
        </Link>
      </div>
    </section>
  )
}
