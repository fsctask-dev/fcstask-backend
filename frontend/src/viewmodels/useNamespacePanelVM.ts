import { namespaceCourses, namespaceUsers, namespaces } from '../mock/data'

export function useNamespacePanelVM(namespaceId?: string) {
  const namespace = namespaces.find((item) => item.id === namespaceId) ?? namespaces[0]

  return {
    namespace,
    users: namespaceUsers,
    courses: namespaceCourses,
  }
}
