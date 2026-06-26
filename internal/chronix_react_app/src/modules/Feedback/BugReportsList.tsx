import React from 'react';
import {FeedbackList} from './FeedbackList';

export const BugReportsList: React.FC = () => {
    return (
        <FeedbackList 
            title="Bug Reports" 
            kind="bug" 
        />
    );
};
