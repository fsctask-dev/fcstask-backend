import { Link } from 'react-router-dom'
import { useNamespacesVM } from '../viewmodels/useNamespacesVM'
import './Pages.css'

export function NamespacesPage() {
  const { namespaces } = useNamespacesVM()

  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Admin</p>
          <h1>Namespaces</h1>
          <p className="subtle">Manage cohorts, courses, and access rights.</p>
        </div>
        <Link className="btn" to="/admin/namespaces/ns-01">
          Open default namespace
        </Link>
      </div>

      <div className="panel">
        <div className="table">
          <div className="table__row table__head">
            <span>ID</span>
            <span>Name</span>
            <span>Slug</span>
            <span>Description</span>
            <span>Gitlab group</span>
            <span>Courses</span>
            <span>Users</span>
          </div>
          {namespaces.map((ns) => (
            <Link key={ns.id} to={`/admin/namespaces/${ns.id}`} className="table__row table__link">
              <span>{ns.id}</span>
              <span>{ns.name}</span>
              <span>{ns.slug}</span>
              <span>{ns.description ?? '—'}</span>
              <span>{ns.gitlabGroupId}</span>
              <span>{ns.coursesCount}</span>
              <span>{ns.usersCount}</span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  )
}
