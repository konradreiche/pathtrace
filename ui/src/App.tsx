import { useState } from "react";
import { Container, Typography } from "@mui/material";
import TaskSelector from "./TaskSelector";
import TaskView from "./TaskView";

export default function App() {
  const [selectedTask, setSelectedTask] = useState<number | null>(null);

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Go Trace Analyzer
      </Typography>

      <TaskSelector onSelect={setSelectedTask} />
      <TaskView taskId={selectedTask} />
    </Container>
  );
}
