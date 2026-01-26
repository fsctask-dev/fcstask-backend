import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { MainLayout } from './layouts/MainLayout'
import { CoursesPage } from './pages/CoursesPage'
import { TasksPage } from './pages/TasksPage'
import { DatabasePage } from './pages/DatabasePage'
import { CreateCoursePage } from './pages/CreateCoursePage'
import { EditCoursePage } from './pages/EditCoursePage'
import { SignupPage } from './pages/SignupPage'
import { SignupFinishPage } from './pages/SignupFinishPage'
import { NamespacesPage } from './pages/NamespacesPage'
import { NamespacePanelPage } from './pages/NamespacePanelPage'
import { InstanceAdminPage } from './pages/InstanceAdminPage'
import { NotReadyPage } from './pages/NotReadyPage'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<MainLayout />}>
          <Route path="/" element={<CoursesPage />} />
          <Route path="/course/create" element={<CreateCoursePage />} />
          <Route path="/course/:courseId" element={<TasksPage />} />
          <Route path="/course/:courseId/database" element={<DatabasePage />} />
          <Route path="/course/:courseId/edit" element={<EditCoursePage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/signup/finish" element={<SignupFinishPage />} />
          <Route path="/admin/namespaces" element={<NamespacesPage />} />
          <Route path="/admin/namespaces/:namespaceId" element={<NamespacePanelPage />} />
          <Route path="/admin/instance" element={<InstanceAdminPage />} />
          <Route path="/not-ready" element={<NotReadyPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
