import React, {useEffect, useRef, useState} from 'react';
import {Alert, Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Typography} from '@mui/material';
import {Close} from '@mui/icons-material';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSlug from 'rehype-slug';
import {apiGet} from '@dsherwin/react-api-interface';

interface HelpDialogProps {
    open: boolean;
    onClose: () => void;
    section?: string; // Optional anchor or heading text to scroll to
}

export const HelpDialog: React.FC<HelpDialogProps> = ({open, onClose, section}) => {
    const [markdown, setMarkdown] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const contentRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (open) {
            const fetchHelp = async () => {
                setLoading(true);
                setError(null);
                try {
                    // Expect JSON response: { "markdown": "..." }
                    const data = await apiGet('/help') as { markdown: string };
                    let content = data.markdown;

                    // If a section is requested, extract only that section
                    if (section) {
                        const lines = content.split('\n');
                        let startIdx = -1;
                        let endIdx = -1;

                        // Look for a header that contains the section name
                        for (let i = 0; i < lines.length; i++) {
                            const line = lines[i].trim();
                            if (line.startsWith('##') && line.toLowerCase().includes(section.toLowerCase())) {
                                startIdx = i;
                                break;
                            }
                        }

                        if (startIdx !== -1) {
                            // Find the end of this section (next major header or horizontal rule)
                            for (let i = startIdx + 1; i < lines.length; i++) {
                                const line = lines[i].trim();
                                if (line.startsWith('---') || (line.startsWith('##') && !line.startsWith('###'))) {
                                    endIdx = i;
                                    break;
                                }
                            }
                            if (endIdx !== -1) {
                                content = lines.slice(startIdx, endIdx).join('\n');
                            } else {
                                content = lines.slice(startIdx).join('\n');
                            }
                        }
                    }

                    setMarkdown(content);
                } catch (e) {
                    console.error('Failed to fetch help markdown', e);
                    setError('Failed to load help documentation.');
                } finally {
                    setLoading(false);
                }
            };
            fetchHelp();
        }
    }, [open, section]);

    useEffect(() => {
        // Only scroll if we are showing the FULL document
        if (!loading && markdown && section && open && !markdown.startsWith('##')) {
            // Give it a moment to render
            setTimeout(() => {
                const id = section.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
                const element = document.getElementById(id);
                if (element) {
                    element.scrollIntoView({behavior: 'smooth'});
                } else {
                  // Try finding by text if ID doesn't match
                  const headings = contentRef.current?.querySelectorAll('h1, h2, h3, h4, h5, h6');
                  headings?.forEach(h => {
                    if (h.textContent?.toLowerCase().includes(section.toLowerCase())) {
                      h.scrollIntoView({behavior: 'smooth'});
                    }
                  });
                }
            }, 500);
        }
    }, [loading, markdown, section, open]);

    return (
        <Dialog 
            open={open} 
            onClose={onClose} 
            maxWidth="md" 
            fullWidth
            scroll="paper"
        >
            <DialogTitle>
                <Box
                    sx={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center"
                    }}>
                    <Typography variant="h6">Chronix Help</Typography>
                    <IconButton onClick={onClose} size="small">
                        <Close />
                    </IconButton>
                </Box>
            </DialogTitle>
            <DialogContent dividers>
                <Box ref={contentRef} sx={{ 
                    '& h1, & h2, & h3, & h4, & h5, & h6': { mt: 3, mb: 2 },
                    '& p, & li': { mb: 1, lineHeight: 1.6 },
                    '& code': { bgcolor: 'action.hover', px: 0.5, borderRadius: 1, fontFamily: 'monospace' },
                    '& pre': { bgcolor: 'action.hover', p: 2, borderRadius: 1, overflow: 'auto', mb: 2 },
                    '& pre code': { bgcolor: 'transparent', px: 0 },
                    '& ul, & ol': { pl: 3, mb: 2 },
                    '& table': { width: '100%', borderCollapse: 'collapse', mb: 2 },
                    '& th, & td': { border: 1, borderColor: 'divider', p: 1, textAlign: 'left' },
                    '& th': { bgcolor: 'action.selected' },
                    '& blockquote': { borderLeft: 4, borderColor: 'primary.main', pl: 2, py: 1, bgcolor: 'action.hover', fontStyle: 'italic', my: 2 },
                    '& img': { maxWidth: '100%' }
                }}>
                    {loading ? (
                        <Box
                            sx={{
                                display: "flex",
                                justifyContent: "center",
                                py: 4
                            }}>
                            <CircularProgress />
                        </Box>
                    ) : error ? (
                        <Alert severity="error">{error}</Alert>
                    ) : (
                        <ReactMarkdown 
                            remarkPlugins={[remarkGfm]} 
                            rehypePlugins={[rehypeSlug]}
                        >
                            {markdown}
                        </ReactMarkdown>
                    )}
                </Box>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} color="primary">Close</Button>
            </DialogActions>
        </Dialog>
    );
};
