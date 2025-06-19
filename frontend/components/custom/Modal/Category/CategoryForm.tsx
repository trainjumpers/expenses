import { LoadingButton } from "@/components/ui/LoadingButton";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import { IconName, IconPicker } from "@/components/ui/icon-picker";
import { Input } from "@/components/ui/input";
import { useState } from "react";

interface CategoryFormProps {
  initialValues: {
    name: string;
    icon: IconName;
  };
  onSubmit: (formData: { name: string; icon: IconName }) => Promise<void>;
  loading: boolean;
  isRefreshing?: boolean;
  submitText: string;
  onOpenChange: (open: boolean) => void;
}

export function CategoryForm({
  initialValues,
  onSubmit,
  loading,
  isRefreshing = false,
  submitText,
  onOpenChange,
}: CategoryFormProps) {
  const [formData, setFormData] = useState(initialValues);

  const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await onSubmit(formData);
  };

  return (
    <form onSubmit={handleFormSubmit}>
      <div className="grid gap-4 py-4">
        <div className="flex items-center gap-4">
          <Input
            id="name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Enter category name"
            className="flex-1"
          />
          <div className="flex items-center gap-2">
            <div>
              <IconPicker
                value={formData.icon || undefined}
                onValueChange={(value) =>
                  setFormData({ ...formData, icon: value })
                }
                defaultValue="circle-dashed"
                modal={true}
              />
            </div>
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={() => onOpenChange(false)}
        >
          Cancel
        </Button>
        <LoadingButton
          type="submit"
          loading={loading || isRefreshing}
          fixedWidth="140px"
          disabled={loading || isRefreshing}
        >
          {submitText}
        </LoadingButton>
      </DialogFooter>
    </form>
  );
}
