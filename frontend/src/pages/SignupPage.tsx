import { Link } from 'react-router-dom'
import { useSignupVM } from '../viewmodels/useSignupVM'
import './Pages.css'

export function SignupPage() {
  const { form, updateField } = useSignupVM()

  return (
    <section className="auth-card">
      <div className="auth-card__header">
        <p className="eyebrow">Invite</p>
        <h1>Join a course</h1>
        <p className="subtle">Enter your invite code and preferred contact details.</p>
      </div>

      <form className="auth-form" onSubmit={(event) => event.preventDefault()}>
        <label>
          Invite code
          <input
            className="input"
            value={form.inviteCode}
            onChange={(event) => updateField('inviteCode', event.target.value)}
          />
        </label>
        <label>
          Email
          <input
            className="input"
            type="email"
            value={form.email}
            onChange={(event) => updateField('email', event.target.value)}
          />
        </label>
        <label>
          Telegram
          <input
            className="input"
            value={form.telegram}
            onChange={(event) => updateField('telegram', event.target.value)}
          />
        </label>
        <label>
          Group
          <input
            className="input"
            value={form.group}
            onChange={(event) => updateField('group', event.target.value)}
          />
        </label>
        <div className="auth-actions">
          <Link className="btn btn-ghost" to="/">
            Cancel
          </Link>
          <Link className="btn" to="/signup/finish">
            Continue
          </Link>
        </div>
      </form>
    </section>
  )
}
