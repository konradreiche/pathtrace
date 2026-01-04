import { useEffect, useState } from "react";
import { FormControl, InputLabel, MenuItem, Select } from "@mui/material";

interface Props {
  onSelect: (id: number) => void;
}

export default function TaskSelector({ onSelect }: Props) {
  const [tasks, setTasks] = useState<number[]>([]);
  const [selectedId, setSelectedId] = useState<number | "">("");

  useEffect(() => {
    fetch("/tasks")
      .then((r) => r.json())
      .then(setTasks)
      .catch(console.error);
  }, []);

  function handleChange(id: number) {
    setSelectedId(id);
    onSelect(id);
  }

  return (
    <FormControl sx={{ minWidth: 300 }}>
      <InputLabel>Task</InputLabel>
      <Select
        value={selectedId}
        label="Task"
        onChange={(e) => handleChange(e.target.value as number)}
      >
        {tasks.map((id) => (
          <MenuItem key={id} value={id}>
            Task {id}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}
