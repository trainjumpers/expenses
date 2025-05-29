import { useUser } from "@/components/custom/Provider/UserProvider";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { useState } from "react";
import { toast } from "sonner";

interface PasswordModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PasswordModal({ isOpen, onOpenChange }: PasswordModalProps) {
  const { updatePassword } = useUser();
  const [formData, setFormData] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      if (formData.newPassword !== formData.confirmPassword) {
        const msg = "Passwordss don't match";
        toast.error(msg);
        throw new Error(msg);
      }

      // Add password strength validation
      if (formData.newPassword.length < 8) {
        const msg = "Password must be at least 8 characters long";
        toast.error(msg);
        throw new Error(msg);
      }
      await updatePassword(formData.currentPassword, formData.newPassword);
      setFormData({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      });
      onOpenChange(false);
    } catch (error) {
      console.error(error);
      // Add user-visible error handling (toast, form error state, etc.)
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Update password</DialogTitle>
          <DialogDescription>
            Enter your current password and choose a new one.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="currentPassword" className="text-right text-sm">
                Current Password
              </Label>
              <Input
                id="currentPassword"
                type="password"
                value={formData.currentPassword}
                onChange={(e) =>
                  setFormData({ ...formData, currentPassword: e.target.value })
                }
                className="col-span-2"
              />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="newPassword" className="text-right text-sm">
                New Password
              </Label>
              <Input
                id="newPassword"
                type="password"
                value={formData.newPassword}
                onChange={(e) =>
                  setFormData({ ...formData, newPassword: e.target.value })
                }
                className="col-span-2"
              />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label htmlFor="confirmPassword" className="text-right text-sm">
                Confirm Password
              </Label>
              <Input
                id="confirmPassword"
                type="password"
                value={formData.confirmPassword}
                onChange={(e) =>
                  setFormData({ ...formData, confirmPassword: e.target.value })
                }
                className="col-span-2"
              />
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
            <Button
              type="submit"
              className="relative min-w-[150px]"
              disabled={isLoading}
            >
              {isLoading ? <Spinner /> : "Update password"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
