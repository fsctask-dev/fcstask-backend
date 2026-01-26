import { useState } from 'react'

export interface SignupFormState {
  inviteCode: string
  email: string
  telegram: string
  group: string
}

export function useSignupVM() {
  const [form, setForm] = useState<SignupFormState>({
    inviteCode: '',
    email: '',
    telegram: '',
    group: '',
  })

  const updateField = <K extends keyof SignupFormState>(key: K, value: SignupFormState[K]) => {
    setForm((current) => ({
      ...current,
      [key]: value,
    }))
  }

  return {
    form,
    updateField,
  }
}
