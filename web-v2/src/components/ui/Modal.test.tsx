import { render, screen, fireEvent } from '@testing-library/react'
import { vi } from 'vitest'
import { Modal } from './Modal'

describe('Modal', () => {
  it('renders nothing when closed', () => {
    render(<Modal open={false} onClose={vi.fn()} title="Test">Content</Modal>)
    expect(screen.queryByText('Test')).not.toBeInTheDocument()
    expect(screen.queryByText('Content')).not.toBeInTheDocument()
  })

  it('renders title and children when open', () => {
    render(<Modal open={true} onClose={vi.fn()} title="My Modal">Modal body</Modal>)
    expect(screen.getByText('My Modal')).toBeInTheDocument()
    expect(screen.getByText('Modal body')).toBeInTheDocument()
  })

  it('renders footer when provided', () => {
    render(
      <Modal open={true} onClose={vi.fn()} title="Test" footer={<button>Save</button>}>
        Content
      </Modal>,
    )
    expect(screen.getByText('Save')).toBeInTheDocument()
  })

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(<Modal open={true} onClose={onClose} title="Test">Content</Modal>)
    fireEvent.click(screen.getByLabelText('Close'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn()
    render(<Modal open={true} onClose={onClose} title="Test">Content</Modal>)
    // The backdrop is the div with bg-black/40
    const backdrop = document.querySelector('.bg-black\\/40')
    expect(backdrop).toBeTruthy()
    fireEvent.click(backdrop!)
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('calls onClose on Escape key', () => {
    const onClose = vi.fn()
    render(<Modal open={true} onClose={onClose} title="Test">Content</Modal>)
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledOnce()
  })
})
