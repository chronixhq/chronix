import {Button} from '@mui/material'
import {render, screen} from '@testing-library/react'
import {describe, expect, it, vi} from 'vitest'
import {WebtaskConnectionEditor} from './WebtaskConnectionEditor.tsx'

describe('WebtaskConnectionEditor', () => {
    it('renders shared actions and treats redacted secrets as blank, non-required inputs', () => {
        render(
            <WebtaskConnectionEditor
                title="Edit Web Task Connection"
                infoTooltip="Update configuration for this API endpoint."
                draft={{
                    name: 'Prod API',
                    authType: 'bearer',
                    authConfig: {token: '<redacted>'},
                    autoCheckEnabled: true,
                    autoCheckSeconds: 300,
                }}
                setDraft={vi.fn()}
                errors={{}}
                agents={[]}
                showPassword={false}
                setShowPassword={vi.fn()}
                testResult={null}
                onDismissTestResult={vi.fn()}
                onTest={vi.fn()}
                onCancel={vi.fn()}
                onSave={vi.fn()}
                saveLabel="Save Changes"
                headerAction={<Button>Duplicate Connection</Button>}
                dangerZone={<div>Danger Zone</div>}
            />
        )

        expect(screen.getByText('Edit Web Task Connection')).toBeInTheDocument()
        expect(screen.getByRole('button', {name: 'Duplicate Connection'})).toBeInTheDocument()
        expect(screen.getByRole('button', {name: 'Save Changes'})).toBeInTheDocument()
        expect(screen.getByText('Danger Zone')).toBeInTheDocument()

        const tokenInput = screen.getByLabelText('Bearer Token') as HTMLInputElement
        expect(tokenInput).toHaveValue('')
        expect(tokenInput).toHaveAttribute('placeholder', '••••••••')
        expect(tokenInput).not.toBeRequired()
    })
})
