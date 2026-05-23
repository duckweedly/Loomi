import { FormEvent, useState } from 'react'
import { Button } from '@lobehub/ui'
import { ArrowUp, Paperclip } from 'lucide-react'

type Props = {
  disabled?: boolean
  placeholder?: string
  onSubmit: (content: string) => void
}

export function Composer({ disabled, placeholder = 'Message Loomi', onSubmit }: Props) {
  const [value, setValue] = useState('')

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const content = value.trim()
    if (!content) return
    onSubmit(content)
    setValue('')
  }

  return (
    <form className="composer glass-panel" onSubmit={handleSubmit}>
      <Button aria-label="Attach" icon={<Paperclip size={15} />} size="small" />
      <textarea
        className="composer-input"
        disabled={disabled}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) handleSubmit(event)
        }}
        placeholder={placeholder}
        rows={1}
        value={value}
      />
      <Button disabled={disabled || value.trim().length === 0} htmlType="submit" icon={<ArrowUp size={15} />} type="primary" />
    </form>
  )
}
