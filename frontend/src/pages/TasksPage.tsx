import { useTasksVM } from '../viewmodels/useTasksVM'
import './Pages.css'

export function TasksPage() {
  const { board, groups, showPastDeadlines, togglePastDeadlines } = useTasksVM()

  return (
    <section className="page-grid">
      <div className="page-header">
        <div>
          <p className="eyebrow">Course</p>
          <h1>{board.courseName}</h1>
          <p className="subtle">Progress overview and upcoming deadlines.</p>
        </div>
        <div className="header-actions">
          <div className="progress-card">
            <div className="progress-card__value">{board.solvedPercent}%</div>
            <div className="progress-card__meta">
              {board.solvedScore}/{board.maxScore} pts
            </div>
          </div>
          <button className="btn btn-ghost" type="button" onClick={togglePastDeadlines}>
            {showPastDeadlines ? 'Hide past deadlines' : 'Show past deadlines'}
          </button>
        </div>
      </div>

      {groups.map((group) => (
        <div key={group.id} className={`panel panel--group ${group.isSpecial ? 'panel--special' : ''}`}>
          <div className="panel__head">
            <div>
              <h2>{group.name}</h2>
              <p className="meta">Score: {group.scoreEarned}/{group.scoreMax}</p>
            </div>
            <div className="deadlines">
              {group.deadlines.map((deadline) => {
                const shouldHide = !showPastDeadlines && deadline.status === 'expired'

                return (
                  <div
                    key={deadline.id}
                    className={`deadline deadline--${deadline.status} ${shouldHide ? 'deadline--hidden' : ''}`}
                  >
                    <div className="deadline__top">
                      <span>{deadline.label}</span>
                      <span>{Math.round(deadline.percent * 100)}%</span>
                    </div>
                    <div className="deadline__time">{new Date(deadline.dueAt).toLocaleString()}</div>
                    <div className="deadline__bar">
                      <span style={{ width: `${deadline.percent * 100}%` }} />
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          <div className="task-grid">
            {group.tasks.map((task) => (
              <article key={task.id} className={`task-card task-card--${task.status}`}>
                <div>
                  <h3>{task.name}</h3>
                  <p className="meta">{task.isSpecial ? 'Special' : task.isBonus ? 'Bonus' : 'Standard'}</p>
                </div>
                <div className="task-card__score">
                  <span>{task.scoreEarned}</span>
                  <small>/{task.score}</small>
                </div>
                <div className="task-card__stat">{task.stats.toFixed(2)} solved</div>
              </article>
            ))}
          </div>
        </div>
      ))}
    </section>
  )
}
