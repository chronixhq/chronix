import React from 'react';
import {FeedbackForm} from './FeedbackForm';
import {useNavigate} from 'react-router';

export const BugReportPage: React.FC = () => {
    const navigate = useNavigate();
    return (
        <FeedbackForm 
            title="Report a Bug" 
            kind="bug" 
            onSubmitSuccess={() => navigate('/')} 
        />
    );
};
