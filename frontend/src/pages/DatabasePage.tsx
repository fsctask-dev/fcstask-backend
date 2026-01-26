import './Pages.css'

const rows = [
  { id: 1, student: 'alex', score: 192, submitted: '2024-10-02' },
  { id: 2, student: 'maria', score: 176, submitted: '2024-10-03' },
  { id: 3, student: 'sasha', score: 144, submitted: '2024-10-05' },
]

export function DatabasePage() {
  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Course</p>
          <h1>All scores</h1>
          <p className="subtle">Snapshot of course-wide submissions.</p>
        </div>
      </div>

      <div className="panel">
        <div className="table">
          <div className="table__row table__head">
            <span>ID</span>
            <span>Student</span>
            <span>Score</span>
            <span>Last submit</span>
          </div>
          {rows.map((row) => (
            <div key={row.id} className="table__row">
              <span>{row.id}</span>
              <span>{row.student}</span>
              <span>{row.score}</span>
              <span>{row.submitted}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
