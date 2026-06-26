import React, {useState} from 'react';
import {Alert, Box, Button, Card, CardContent, Divider, IconButton, List, ListItem, ListItemText, TextField, Typography} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import {VStack} from "@dsherwin/mui-kit";
import {apiPost} from '@dsherwin/react-api-interface';
import {formatAPIError} from "../../lib/errors";

interface FeedbackFormProps {
    title: string;
    kind: 'bug' | 'feature';
    onSubmitSuccess?: () => void;
}

export const FeedbackForm: React.FC<FeedbackFormProps> = ({title, kind, onSubmitSuccess}) => {
    const [summary, setSummary] = useState('');
    const [description, setDescription] = useState('');
    const [files, setFiles] = useState<File[]>([]);
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);
    const [dragOver, setDragOver] = useState(false);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files) {
            const selectedFiles = Array.from(e.target.files);
            setFiles(prev => [...prev, ...selectedFiles]);
            e.target.value = '';
        }
    };

    const onDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setDragOver(true);
    };

    const onDragLeave = () => {
        setDragOver(false);
    };

    const onDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setDragOver(false);
        if (e.dataTransfer.files) {
            const droppedFiles = Array.from(e.dataTransfer.files);
            setFiles(prev => [...prev, ...droppedFiles]);
        }
    };

    const removeFile = (index: number) => {
        setFiles(prev => prev.filter((_, i) => i !== index));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitting(true);
        setError(null);
        setSuccess(false);

        const formData = new FormData();
        formData.append('summary', summary);
        formData.append('description', description);
        files.forEach((file) => {
            formData.append('attachments', file);
        });

        const endpoint = kind === 'bug' ? '/feedback/bug-report' : '/feedback/feature-request';

        try {
            const res = await apiPost(endpoint, formData) as any;
            if (res.status === 'ok' || res.id) {
                setSuccess(true);
                setSummary('');
                setDescription('');
                setFiles([]);
                if (onSubmitSuccess) {
                    setTimeout(onSubmitSuccess, 2000);
                }
            } else {
                setError(res.message || 'Failed to submit feedback');
            }
        } catch (e: any) {
            console.error(e);
            setError(formatAPIError(e, 'Failed to submit feedback.'));
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <Box sx={{px: {xs: 1, md: 2}, py: 2, width: '100%'}}>
            <VStack spacing={3} sx={{maxWidth: 800, width: '100%', mx: 'auto'}}>
                <Typography variant="h5">{title}</Typography>
                <Divider/>

                {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}
                {success && (
                    <Alert severity="success" onClose={() => setSuccess(false)}>
                        Thank you! Your {kind === 'bug' ? 'bug report' : 'feature request'} has been submitted.
                    </Alert>
                )}

                <Card variant="outlined" sx={{borderRadius: 3}}>
                    <CardContent>
                        <form onSubmit={handleSubmit}>
                            <VStack spacing={3}>
                                <TextField
                                    fullWidth
                                    label="Summary"
                                    required
                                    value={summary}
                                    onChange={(e) => setSummary(e.target.value)}
                                    placeholder={kind === 'bug' ? "Briefly describe the bug" : "Short title for your request"}
                                />
                                <TextField
                                    fullWidth
                                    label="Description"
                                    required
                                    multiline
                                    rows={6}
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    placeholder={kind === 'bug' ? "Steps to reproduce, expected vs actual behavior..." : "Describe the feature and why it would be useful"}
                                />
                                <Box>
                                    <Typography variant="subtitle2" gutterBottom>Attachments (Images/Documents)</Typography>
                                    <Box
                                        onDragOver={onDragOver}
                                        onDragLeave={onDragLeave}
                                        onDrop={onDrop}
                                        onClick={() => document.getElementById('file-input')?.click()}
                                        sx={{
                                            border: '2px dashed',
                                            borderColor: dragOver ? 'primary.main' : 'divider',
                                            borderRadius: 2,
                                            p: 3,
                                            textAlign: 'center',
                                            cursor: 'pointer',
                                            bgcolor: dragOver ? 'action.hover' : 'background.paper',
                                            transition: 'all 0.2s ease-in-out',
                                            '&:hover': {
                                                borderColor: 'primary.main',
                                                bgcolor: 'action.hover',
                                            }
                                        }}
                                    >
                                        <VStack spacing={1} alignItems="center">
                                            <CloudUploadIcon color="primary" sx={{ fontSize: 40 }} />
                                            <Typography variant="body2">
                                                Drag and drop files here, or click to select
                                            </Typography>
                                            <Typography variant="caption" sx={{
                                                color: "text.secondary"
                                            }}>
                                                Images, PDFs, logs, etc.
                                            </Typography>
                                        </VStack>
                                        <input
                                            id="file-input"
                                            type="file"
                                            multiple
                                            onChange={handleFileChange}
                                            style={{ display: 'none' }}
                                        />
                                    </Box>

                                    {files.length > 0 && (
                                        <List dense sx={{ mt: 1, border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
                                            {files.map((file, index) => (
                                                <ListItem
                                                    key={index}
                                                    secondaryAction={
                                                        <IconButton edge="end" aria-label="delete" onClick={() => removeFile(index)}>
                                                            <DeleteIcon fontSize="small" />
                                                        </IconButton>
                                                    }
                                                    divider={index < files.length - 1}
                                                >
                                                    <ListItemText
                                                        primary={file.name}
                                                        secondary={`${(file.size / 1024).toFixed(1)} KB`}
                                                        slotProps={{ primary: { variant: 'body2', noWrap: true } }}
                                                    />
                                                </ListItem>
                                            ))}
                                        </List>
                                    )}
                                </Box>
                                <Box sx={{pt: 1}}>
                                    <Button
                                        type="submit"
                                        variant="contained"
                                        disabled={submitting}
                                    >
                                        {submitting ? 'Submitting...' : `Submit ${kind === 'bug' ? 'Bug Report' : 'Feature Request'}`}
                                    </Button>
                                </Box>
                            </VStack>
                        </form>
                    </CardContent>
                </Card>
            </VStack>
        </Box>
    );
};
