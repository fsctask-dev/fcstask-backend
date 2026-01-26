import { Link, useParams } from 'react-router-dom'
import { useNamespacePanelVM } from '../viewmodels/useNamespacePanelVM'
import './Pages.css'

export function NamespacePanelPage() {
  const { namespaceId } = useParams()
  const { namespace, users, courses } = useNamespacePanelVM(namespaceId)

  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Namespace</p>
          <h1>{namespace.name}</h1>
          <p className="subtle">{namespace.description}</p>
          <p className="meta">Slug: {namespace.slug} · ID: {namespace.id} · GitLab: {namespace.gitlabGroupId}</p>
        </div>
        <Link className="btn btn-ghost" to="/admin/namespaces">
          Back to list
        </Link>
      </div>

      <div className="panel">
        <div className="panel__head">
          <h2>Namespace users</h2>
          <button className="btn btn-ghost" type="button">
            Add user
          </button>
        </div>
        <div className="table">
          <div className="table__row table__head">
            <span>ID</span>
            <span>Username</span>
            <span>RMS ID</span>
            <span>Role</span>
            <span>Change role</span>
          </div>
          {users.map((user) => (
            <div key={user.id} className="table__row">
              <span>{user.id}</span>
              <span>{user.username}</span>
              <span>{user.rmsId}</span>
              <span>{user.role.replace('_', ' ')}</span>
              <select className="input input--small" defaultValue={user.role}>
                <option value="namespace_admin">Namespace admin</option>
                <option value="program_manager">Program manager</option>
                <option value="student">Student</option>
              </select>
            </div>
          ))}
        </div>
      </div>

      <div className="panel">
        <div className="panel__head">
          <h2>Namespace courses</h2>
          <button className="btn" type="button">
            Create course
          </button>
        </div>
        <div className="table">
          <div className="table__row table__head">
            <span>ID</span>
            <span>Course</span>
            <span>Status</span>
            <span>Gitlab group</span>
            <span>Owners</span>
          </div>
          {courses.map((course) => (
            <Link key={course.id} to={course.url} className="table__row table__link">
              <span>{course.id}</span>
              <span>{course.name}</span>
              <span>{course.status}</span>
              <span>{course.gitlabGroup}</span>
              <span>{course.owners.join(', ')}</span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  )
}
