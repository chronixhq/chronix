import '@testing-library/jest-dom/vitest'
import React from 'react'
import {vi} from 'vitest'

function createStack(children: React.ReactNode, props: Record<string, unknown>) {
    const {
        alignItems,
        direction,
        gap,
        justifyContent,
        spacing,
        sx,
        ...domProps
    } = props
    void alignItems
    void direction
    void gap
    void justifyContent
    void spacing
    void sx
    return React.createElement('div', domProps, children)
}

vi.mock('@dsherwin/mui-kit', async () => {
    const MuiKit = ({children}: { children: React.ReactNode }) => React.createElement(React.Fragment, null, children)
    const HStack = ({children, ...props}: React.HTMLAttributes<HTMLDivElement>) => createStack(children, props as Record<string, unknown>)
    const VStack = ({children, ...props}: React.HTMLAttributes<HTMLDivElement>) => createStack(children, props as Record<string, unknown>)
    const PasswordTextField = React.forwardRef<HTMLInputElement, Record<string, unknown>>((props, ref) => (
        React.createElement('input', {...props, ref})
    ))
    const ConstrainedMainContent = ({children}: { children: React.ReactNode }) => React.createElement(React.Fragment, null, children)

    return {
        MuiKit,
        Messages: MuiKit,
        MuiPrompts: MuiKit,
        HStack,
        VStack,
        PasswordTextField,
        ConstrainedMainContent,
        SelectWithLabel: (props: Record<string, unknown>) => React.createElement('div', props),
        SwitchWithLabel: (props: Record<string, unknown>) => React.createElement('div', props),
        useMessagesContext: () => ({
            showSuccess: vi.fn(),
            showError: vi.fn(),
            showInfo: vi.fn(),
            showWarning: vi.fn(),
        }),
        useMuiPrompts: () => ({
            confirmPrompt: vi.fn().mockResolvedValue(true),
            textPrompt: vi.fn().mockResolvedValue(''),
            alert: vi.fn(),
        }),
    }
})
