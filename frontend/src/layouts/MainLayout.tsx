import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { coursesShort, currentUser } from '../mock/data'
import './MainLayout.css'

export function MainLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const isCourseRoute = location.pathname.startsWith('/course/')
  const isSignup = location.pathname.startsWith('/signup')
  const courseBase = isCourseRoute ? location.pathname.split('/').slice(0, 3).join('/') : ''

  if (isSignup) {
    return (
      <div className="shell shell--auth">
        <Outlet />
      </div>
    )
  }

  return (
    <div className="shell">
      <header className="topbar">
        <div className="topbar__brand">
          <Link to="/" className="brand">
            <span className="brand__mark">MT</span>
            <span className="brand__title">FCS Task</span>
          </Link>
        </div>

        <nav className="topbar__nav">
          {isCourseRoute ? (
            <>
              <NavLink to={courseBase} className="nav-link">
                Assignments
              </NavLink>
              <a className="nav-link" href="https://gitlab.com" target="_blank" rel="noreferrer">
                My Repo
              </a>
              <a className="nav-link" href="https://gitlab.com" target="_blank" rel="noreferrer">
                My Submits
              </a>
              <NavLink to={`${courseBase}/database`} className="nav-link">
                All Scores
              </NavLink>
              <NavLink to={`${courseBase}/edit`} className="nav-link">
                Edit Course
              </NavLink>
              <NavLink to="/admin/namespaces" className="nav-link">
                Namespaces
              </NavLink>
            </>
          ) : (
            <>
              <NavLink to="/" className="nav-link">
                Courses
              </NavLink>
              <NavLink to="/admin/namespaces" className="nav-link">
                Namespaces
              </NavLink>
              <NavLink to="/admin/instance" className="nav-link">
                Instance Panel
              </NavLink>
            </>
          )}
        </nav>

        <div className="topbar__actions">
          {isCourseRoute && (
            <div className="course-switch">
              <span className="course-switch__label">Course</span>
              <div className="course-switch__control">
                <select
                  className="course-switch__select"
                  defaultValue={coursesShort[0]?.url}
                  onChange={(event) => navigate(event.target.value)}
                >
                  {coursesShort.map((course) => (
                    <option key={course.url} value={course.url}>
                      {course.name}
                    </option>
                  ))}
                </select>
                <span className="course-switch__chevron" aria-hidden="true">
                  ▾
                </span>
              </div>
            </div>
          )}

          <div className="user-chip">
            <span className="user-chip__initials">{currentUser.initials}</span>
            <div>
              <div className="user-chip__name">{currentUser.username}</div>
              <div className="user-chip__role">{currentUser.role.replace('_', ' ')}</div>
            </div>
          </div>
        </div>
      </header>

      <main className="page">
        <Outlet />
      </main>
    </div>
  )
}
