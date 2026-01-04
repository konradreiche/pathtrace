import { useEffect, useState } from "react";
import { Typography } from "@mui/material";

interface Props {
  taskId: number | null;
}

export default function TaskView({ taskId }: Props) {
  const [nodes, setNodes] = useState<unknown[]>([]);

  useEffect(() => {
    if (taskId === null) return;
    fetch(`/task/${taskId}`)
      .then((r) => r.json())
      .then((data) => {
        console.log(data);
        setNodes(data);
      })
      .catch(console.error);
  }, [taskId]);

  if (taskId === null) return null;

  return (
    <Typography sx={{ mt: 2 }}>
      Task {taskId} — {nodes.length} nodes
    </Typography>
  );
}
